package object

type ContextKey string

const RecMsgSecondTimeKey = ContextKey("receive message timestamp in second")
const TraceId = ContextKey("service trace id")
