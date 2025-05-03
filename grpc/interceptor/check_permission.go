package interceptor

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/nioliu/commons/errs"
	"github.com/nioliu/commons/grpc/object"
	"github.com/nioliu/commons/log"
	"github.com/nioliu/protocols/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// protectedEndpoints stores the mapping of service/method to required permissions
	protectedEndpoints = make(map[string][]string)
	configMutex        sync.RWMutex
	userServiceClient  user.UserServiceClient
)

func init() {
	// Initialize user service client
	if err := InitUserServiceClient(); err != nil {
		panic(fmt.Sprintf("failed to initialize user service client: %v", err))
	}

	// Load permission configuration
	configPath := os.Getenv("PERMISSION_CONFIG_PATH")
	if configPath == "" {
		configPath = "/conf/permission.config" // default path
	}

	if err := LoadPermissionConfig(configPath); err != nil {
		panic(fmt.Sprintf("failed to load permission config: %v", err))
	}
}

// LoadPermissionConfig loads the permission configuration from a file
func LoadPermissionConfig(configPath string) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read permission config: %v", err)
	}

	// Parse the config file (assuming it's in format: service/method=permission1,permission2)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		endpoint := strings.TrimSpace(parts[0])
		permissions := strings.Split(parts[1], ",")
		for i := range permissions {
			permissions[i] = strings.TrimSpace(permissions[i])
		}
		protectedEndpoints[endpoint] = permissions
	}

	return nil
}

// InitUserServiceClient initializes the user service client
func InitUserServiceClient() error {
	target := os.Getenv("PERMISSION_CHECK_USER_SERVICE_TARGET")
	if target == "" {
		target = "user-service:8000"
	}
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(GetBackCallLogFunc()))
	if err != nil {
		return fmt.Errorf("failed to connect to user service: %v", err)
	}
	userServiceClient = user.NewUserServiceClient(conn)
	return nil
}

// GetCheckPermissionFunc returns a gRPC interceptor that checks permissions
func GetCheckPermissionFunc() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		// Get the full method name (e.g., "/service.UserService/QueryUserInfo")
		fullMethod := info.FullMethod

		// Check if this endpoint requires permission checking
		configMutex.RLock()
		requiredPermissions, needsCheck := protectedEndpoints[fullMethod]
		configMutex.RUnlock()

		if !needsCheck {
			return handler(ctx, req)
		}

		// Get user ID from context
		userID, err := object.GetUserIDFromCtx(ctx)
		if err != nil {
			log.ErrorWithCtxFields(ctx, "failed to get user ID from context", zap.Error(err))
			return nil, errs.NewError(0, "unauthorized: missing user ID")
		}

		// Use reflection to find user ID field in request
		reqValue := reflect.ValueOf(req)
		if reqValue.Kind() == reflect.Ptr {
			reqValue = reqValue.Elem()
		}

		// Look for common user ID field names
		userIDFields := []string{"UserId", "UserID", "Userid", "user_id"}
		var requestUserID string
		for _, fieldName := range userIDFields {
			field := reqValue.FieldByName(fieldName)
			if field.IsValid() && field.Kind() == reflect.String {
				requestUserID = field.String()
				break
			}
		}

		// If we found a user ID in the request, check if it matches the context
		if requestUserID != "" && requestUserID != userID {
			// Check if user has required permissions
			hasPermission := false
			for _, permission := range requiredPermissions {
				// Call user-service to check permission
				rsp, err := userServiceClient.CheckUserPermission(ctx, &user.CheckUserPermissionReq{
					UserId:       userID,
					TargetUserId: requestUserID,
					Permission:   user.SubAccountPermissionType(user.SubAccountPermissionType_value[permission]),
				})
				if err != nil {
					log.ErrorWithCtxFields(ctx, "failed to check permission", zap.Error(err))
					return nil, errs.NewError(0, "failed to check permission")
				}

				if rsp.HasPermission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				log.ErrorWithCtxFields(ctx, "user does not have required permission",
					zap.String("user_id", userID),
					zap.String("request_user_id", requestUserID),
					zap.Strings("required_permissions", requiredPermissions))
				return nil, errs.NewError(0, "unauthorized: insufficient permissions")
			}
		}

		// If we get here, either:
		// 1. No user ID field found in request (internal service call)
		// 2. User ID matches context
		// 3. User has required permissions
		return handler(ctx, req)
	}
}
