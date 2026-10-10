package runtime

import "mk/internal/oop"

func NewRequestContext() *oop.MapObj {
	root := oop.NewMap()
	header := oop.NewMap()
	cookie := oop.NewMap()
	query := oop.NewMap()
	body := oop.NewMap()
	remote := oop.NewMap()
	ctx := oop.NewMap()

	root.PutString("method", oop.NewString("GET"))
	root.PutString("path", oop.NewString("/"))
	root.PutString("header", header)
	root.PutString("cookie", cookie)
	root.PutString("query", query)
	root.PutString("body", body)
	root.PutString("remote", remote)
	root.PutString("context", ctx)

	root.PutString("url", oop.NewMap())
	header.PutString("authorization", oop.NewString(""))
	cookie.PutString("session_id", oop.NewString(""))
	query.PutString("user_id", oop.NewString("0"))
	remote.PutString("ip", oop.NewString("127.0.0.1"))
	remote.PutString("port", oop.NewInt(0))
	ctx.PutString("tenant", oop.NewString("default"))

	return root
}
