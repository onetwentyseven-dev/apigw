package apigw

type Middleware func(Handler) Handler
type WsMiddleware func(WsHandler) WsHandler

func UseMiddleware(handler Handler, wares ...Middleware) Handler {

	if len(wares) == 0 {
		return handler
	}

	for i := len(wares) - 1; i >= 0; i-- {
		handler = wares[i](handler)
	}

	return handler

}

func UseWsMiddleware(handler WsHandler, wares ...WsMiddleware) WsHandler {

	if len(wares) == 0 {
		return handler
	}

	for i := len(wares) - 1; i >= 0; i-- {
		handler = wares[i](handler)
	}

	return handler

}
