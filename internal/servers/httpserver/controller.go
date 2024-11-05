package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/fydmer/fileserver/internal/domain/service"
	"github.com/fydmer/fileserver/pkg/random"
)

const maxFileSize = 10 * 1024 * 1024 * 1024

type controllerHandler struct {
	controller service.Controller
}

type ControllerServer struct {
	server  *http.Server
	handler *controllerHandler
}

func RunControllerServer(ctx context.Context, port int, controller service.Controller) (*ControllerServer, error) {
	handler := &controllerHandler{
		controller: controller,
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", handler.getEndpoints()))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "http server error", slog.String("error", err.Error()))
		}
	}()

	slog.InfoContext(ctx, "http server started", slog.Int("port", port))

	return &ControllerServer{
		server:  server,
		handler: handler,
	}, nil
}

func (s *ControllerServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = s.server.Shutdown(ctx)
}

func (x *controllerHandler) getEndpoints() *http.ServeMux {
	mux := http.NewServeMux()
	{
		mux.HandleFunc("POST /nodes", x.createNode)
		mux.HandleFunc("POST /files", x.uploadFile)
		mux.HandleFunc("GET /files/{location}", x.downloadFile)
		mux.HandleFunc("DELETE /files/{location}", x.deleteFile)

		mux.HandleFunc("GET /tools/file-generator", x.generateFile)
	}
	return mux
}

func (x *controllerHandler) createNode(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addr := r.Form.Get("addr")

	joinNode, err := x.controller.JoinNode(r.Context(), &service.ControllerJoinNodeIn{Addr: addr})
	if err != nil {
		httpError(w, err.Error(), getCodeFromPayloadError(err))
		return
	}

	httpJson(w, map[string]any{
		"node_id": joinNode.Id,
	}, http.StatusOK)
	return
}

func (x *controllerHandler) uploadFile(w http.ResponseWriter, r *http.Request) {
	contentLengthStr := r.Header.Get("Content-Length")
	contentLength, _ := strconv.ParseInt(contentLengthStr, 10, 64)
	if contentLengthStr == "" || contentLength == 0 {
		httpError(w, "header 'Content-Length' is required", http.StatusLengthRequired)
		return
	}

	if contentLength > maxFileSize {
		httpError(w, "Content length is too large", http.StatusRequestEntityTooLarge)
		return
	}

	location := parseContentDisposition(r.Header.Get("Content-Disposition"))
	if location == "" {
		httpError(w, "header 'Content-Disposition' is required", http.StatusPreconditionFailed)
		return
	}

	fileDir, fileName := path.Split(location)
	if fileDir != "" {
		httpError(w, "unsupported file directories (`/`)", http.StatusBadRequest)
		return
	}

	uploadedFile, err := x.controller.UploadFile(r.Context(), &service.ControllerUploadFileIn{
		Location: fileName,
		Size:     contentLength,
		Content:  r.Body,
	})
	if err != nil {
		httpError(w, err.Error(), getCodeFromPayloadError(err))
		return
	}

	httpJson(w, map[string]any{
		"file_id":  uploadedFile.Id,
		"location": fileName,
		"size":     contentLength,
	}, http.StatusCreated)
	return
}

func (x *controllerHandler) downloadFile(w http.ResponseWriter, r *http.Request) {
	location := r.PathValue("location")

	searchFile, err := x.controller.SearchFile(r.Context(), &service.ControllerSearchFileIn{
		Location: location,
	})
	if err != nil {
		httpError(w, err.Error(), getCodeFromPayloadError(err))
		return
	}

	if searchFile.Status != 0 {
		httpError(w, fmt.Sprintf("unable to download file. the file status is %d. "+
			"you could remove and then upload this file again", searchFile.Status),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", searchFile.Size))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(location)))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = x.controller.DownloadFile(r.Context(), &service.ControllerDownloadFileIn{
		Id:      searchFile.Id,
		Content: w,
	})
	if err != nil {
		httpError(w, err.Error(), getCodeFromPayloadError(err))
		return
	}

	return
}

func httpOkIfNotFoundCode(w http.ResponseWriter, err error) {
	if code := getCodeFromPayloadError(err); code != http.StatusNotFound {
		httpError(w, err.Error(), code)
		return
	}
	return
}

func (x *controllerHandler) deleteFile(w http.ResponseWriter, r *http.Request) {
	location := r.PathValue("location")

	searchFile, err := x.controller.SearchFile(r.Context(), &service.ControllerSearchFileIn{
		Location: location,
	})
	if err != nil {
		httpOkIfNotFoundCode(w, err)
		return
	}

	_, err = x.controller.DeleteFile(r.Context(), &service.ControllerDeleteFileIn{
		Id: searchFile.Id,
	})
	if err != nil {
		httpOkIfNotFoundCode(w, err)
		return
	}
	return
}

func (x *controllerHandler) generateFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	size, err := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
	if err != nil {
		httpError(w, fmt.Sprintf("failed to parse 'size' value: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(name)))
	w.Header().Set("Content-Type", "application/octet-stream")

	if err = random.WriteRandomData(r.Context(), w, size); err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
