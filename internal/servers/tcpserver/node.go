package tcpserver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"github.com/fydmer/fileserver/internal/domain/service"
)

type nodeHandler struct {
	node service.Node
}

type NodeServer struct {
	listener net.Listener
	handler  *nodeHandler
}

func RunNodeServer(ctx context.Context, port int, node service.Node) (*NodeServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	handler := &nodeHandler{node: node}

	server := &NodeServer{
		listener: listener,
		handler:  handler,
	}

	go func(lis net.Listener) {
		for {
			conn, _ := lis.Accept()

			select {
			case <-ctx.Done():
				_ = lis.Close()
				return
			default:
				if conn != nil {
					go server.router(context.Background(), conn)
				}
			}
		}
	}(listener)

	slog.InfoContext(ctx, "tcp server started", slog.Int("port", port))

	return server, nil
}

func (s *NodeServer) Close() {
	_ = s.listener.Close()
}

func (x *nodeHandler) saveFile(ctx context.Context, r *bufio.Reader, w *bufio.Writer, headerValue string) error {
	sp := strings.SplitN(headerValue, ":", 2)
	if len(sp) != 2 {
		return fmt.Errorf("invalid header: %s", headerValue)
	}

	size, err := strconv.ParseInt(sp[0], 10, 64)
	if err != nil {
		return err
	}

	saveFile, err := x.node.SaveFile(ctx, &service.NodeSaveFileIn{
		Name:       sp[1],
		DataReader: io.LimitReader(r, size),
	})
	if err != nil {
		return err
	}

	_, err = w.WriteString(fmt.Sprintf("%d\n", saveFile.Written))
	if err != nil {
		return err
	}

	return nil
}

func (x *nodeHandler) getFile(ctx context.Context, r *bufio.Reader, w *bufio.Writer, filename string) error {
	getFile, err := x.node.GetFile(ctx, &service.NodeGetFileIn{
		Name:       filename,
		DataWriter: w,
	})
	if err != nil {
		return err
	}

	if err = w.Flush(); err != nil {
		return err
	}

	writtenStr, err := r.ReadString('\n')
	if err != nil {
		return err
	}

	written, err := strconv.ParseInt(strings.TrimSpace(writtenStr), 10, 64)
	if err != nil {
		return err
	}

	if getFile.Written != written {
		return fmt.Errorf("file size not equal %d != %d", getFile.Written, written)
	}

	return nil
}

func (x *nodeHandler) deleteFile(ctx context.Context, _ *bufio.Reader, _ *bufio.Writer, filename string) error {
	_, err := x.node.DeleteFile(ctx, &service.NodeDeleteFileIn{
		Name: filename,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *NodeServer) router(parentCtx context.Context, conn net.Conn) {
	defer conn.Close()

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	r, w := bufio.NewReader(conn), bufio.NewWriter(conn)
	defer w.Flush()

	headerStr, err := r.ReadString('\n')
	if err != nil {
		return
	}
	headerStr = strings.TrimSpace(headerStr)

	switch prefix, headerValue := parseNodeHeader(headerStr); prefix {
	case "save_file":
		err = s.handler.saveFile(ctx, r, w, headerValue)
	case "get_file":
		err = s.handler.getFile(ctx, r, w, headerValue)
	case "delete_file":
		err = s.handler.deleteFile(ctx, r, w, headerValue)
	default:
	}

	if err != nil {
		slog.Error("operation was fatal", slog.String("error", err.Error()))

		errMsg := strings.ReplaceAll(err.Error(), "\n", " ")
		_, _ = w.WriteString(fmt.Sprintf("error:%s\n", errMsg))
	}
	return
}
