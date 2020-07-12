package http

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/gorilla/mux"
	"github.com/iqdf/pastebin-service/domain"
)

// PasteData is the payload data related to /paste endpoints
// from the request body, request form and also the response back.
// When parsing from request payload, it can take it from
// request body or form depending on the Content-Type of the request.
type PasteData struct {
	Title       string        `json:"paste_title" form:"paste_title"`
	TextData    string        `json:"paste_code" form:"paste_code"`
	StorageURL  string        `json:"storage_url,omitempty" form:"storage_url,omitempty"`
	Private     bool          `json:"paste_private" form:"paste_private,omitempty"`
	PasteExpiry time.Duration `json:"paste_expiry,omitempty" form:"paste_expiry,omitempty"`
}

// PasteResponse is the response structure related to /paste endpoints
// which contains the payload/data and the message to client
type PasteResponse struct {
	Data    *PasteData `json:"data,omitempty"`
	Message string     `json:"message,omitempty"`
}

// PasteHandler handles incoming request and call service
// to process some result before returning the response to client
type PasteHandler struct {
	service domain.PasteService
}

// NewPasteHandler creates new PastHandler using given
// a paste-related service
func NewPasteHandler(service domain.PasteService) *PasteHandler {
	handler := &PasteHandler{
		service: service,
	}
	return handler
}

// Paste transforms current paste data from request payload
// to data transfer object (domain.Paste)
func (data PasteData) Paste() domain.Paste {
	return domain.Paste{
		Title:      data.Title,
		TextData:   data.TextData,
		Private:    data.Private,
		ExpiredAt:  time.Now().Add(data.PasteExpiry),
		StorageURL: data.StorageURL,
	}
}

func makePasteData(paste domain.Paste) PasteData {
	return PasteData{
		Title:       paste.Title,
		TextData:    paste.TextData,
		StorageURL:  paste.StorageURL,
		PasteExpiry: time.Until(paste.ExpiredAt),
		Private:     paste.Private,
	}
}

// Routes register each handle function to a endpoint path url
func (handler *PasteHandler) Routes(router *mux.Router, middleware mux.MiddlewareFunc) {
	// register middleware and handler method here ...
	router.Handle("/{shortURLPath}", middleware(handler.getPasteHandlerFunc())).
		Methods("GET").Name("READ_PASTE")
	router.Handle("/", middleware(handler.createPasteHandlerFunc())).
		Methods("POST").Name("WRITE_PASTE")

	return
}

func (handler *PasteHandler) getPasteHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse url query parameter to get paste shortURLPath
		params := mux.Vars(r)
		shortURLPath := params["shortURLPath"]

		// retrieve paste text data and other information
		paste, err := handler.service.ReadPaste(r.Context(), shortURLPath)

		// prepare response payload and HTTP Status
		pasteData := makePasteData(paste)
		response := makeResponse(&pasteData, err)
		httpStatus := getHTTPStatus(err)

		// write payload and http status into http.ResponseWriter
		writeResponse(w, httpStatus, response)
	}
}

func (handler *PasteHandler) createPasteHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handling logic here ...
		r.ParseForm()

		var pasteData PasteData

		err := parseRequestData(r, &pasteData)
		if err != nil {
			// handle form parsing error due to invalid request
			// return HTTP 400 BAD_REQUEST
			response := makeResponse(nil, err)
			writeResponse(w, http.StatusBadRequest, response)
			return
		}

		// convert paste request data into data transfer object (domain.Paste)
		paste := pasteData.Paste()
		paste.AuthorUserID = "anonymous3887" // TODO: add this userID from Auth or API Key
		shortURLPath, err := handler.service.WritePaste(r.Context(), paste)

		// preserve or add leading slash
		// but trim trailing slash if any
		if !strings.HasPrefix(shortURLPath, "/") {
			shortURLPath = "/" + shortURLPath
		}
		strings.TrimRight(shortURLPath, "/")

		// perform redirect to new paste url
		http.Redirect(w, r, shortURLPath, http.StatusFound)
	}
}

func parseRequestData(r *http.Request, pasteData *PasteData) error {
	if hasContentType(r, "application/json") {
		json.NewDecoder(r.Body).Decode(pasteData)
		return nil
	} else if hasContentType(r, "application/x-www-form-urlencoded") {
		form.NewDecoder().Decode(pasteData, r.Form)
		return nil
	}
	return fmt.Errorf("Content Type not supported")
}

func hasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}

func makeResponse(pasteData *PasteData, err error) PasteResponse {
	if err != nil {
		return PasteResponse{Message: err.Error()}
	}
	return PasteResponse{
		Data:    pasteData,
		Message: "Success",
	}
}

func getHTTPStatus(err error) int {
	// TODO: logic to decide/swicth status based on error
	return http.StatusOK
}

func writeResponse(w http.ResponseWriter, httpStatus int, resp PasteResponse) {
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}
