package gqlclient

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
)

// Upload is a file upload.
//
// See the GraphQL multipart request specification for details:
// https://github.com/jaydenseric/graphql-multipart-request-spec
type Upload struct {
	Filename string
	MIMEType string
	Body     io.Reader
}

// MarshalJSON implements json.Marshaler.
func (Upload) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

func writeMultipart(pw *io.PipeWriter, uploads map[string]Upload, operations io.Reader) (contentType string) {
	mw := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer mw.Close()

		mapData := make(map[string][]string)
		for k := range uploads {
			mapData[k] = []string{"variables." + k}
		}

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="operations"`)
		h.Set("Content-Type", "application/json")
		w, err := mw.CreatePart(h)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create operations part: %v", err))
			return
		}
		if _, err := io.Copy(w, operations); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to write operations part: %v", err))
			return
		}

		h = make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="map"`)
		h.Set("Content-Type", "application/json")
		w, err = mw.CreatePart(h)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create map part: %v", err))
			return
		}
		if err := json.NewEncoder(w).Encode(mapData); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to write map part: %v", err))
			return
		}

		for k, upload := range uploads {
			dispParams := map[string]string{"name": k}
			if upload.Filename != "" {
				dispParams["filename"] = upload.Filename
			}

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", mime.FormatMediaType("form-data", dispParams))
			if upload.MIMEType != "" {
				h.Set("Content-Type", upload.MIMEType)
			}

			w, err := mw.CreatePart(h)
			if err != nil {
				pw.CloseWithError(fmt.Errorf("failed to create upload %q part: %v", k, err))
				return
			}
			if _, err := io.Copy(w, upload.Body); err != nil {
				pw.CloseWithError(fmt.Errorf("failed to write upload %q part: %v", k, err))
				return
			}
		}
	}()

	return mw.FormDataContentType()
}
