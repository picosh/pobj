package pobj

import (
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/picosh/send/auth"
	"github.com/picosh/send/list"
	"github.com/picosh/send/pipe"
	wishrsync "github.com/picosh/send/protocols/rsync"
	"github.com/picosh/send/protocols/scp"
	"github.com/picosh/send/protocols/sftp"
	"github.com/picosh/send/proxy"
	"github.com/picosh/send/utils"
)

func createRouter(handler utils.CopyFromClientHandler) proxy.Router {
	return func(sh ssh.Handler, s ssh.Session) []wish.Middleware {
		return []wish.Middleware{
			pipe.Middleware(handler, ""),
			list.Middleware(handler),
			scp.Middleware(handler),
			wishrsync.Middleware(handler),
			auth.Middleware(handler),
			lm.Middleware(),
		}
	}
}

func WithProxy(handler utils.CopyFromClientHandler, otherMiddleware ...wish.Middleware) ssh.Option {
	return func(server *ssh.Server) error {
		err := sftp.SSHOption(handler)(server)
		if err != nil {
			return err
		}

		return proxy.WithProxy(createRouter(handler), otherMiddleware...)(server)
	}
}
