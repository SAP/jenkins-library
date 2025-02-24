package gcs

import "github.com/SAP/jenkins-library/pkg/log"

// progressWriter tracks progress and updates a progress bar.
type progressWriter struct {
	Total      int64
	Uploaded   int64
	OnProgress func(uploaded, total int64)
}

func newProgressW(fileSize int64) *progressWriter {
	return &progressWriter{
		Total: fileSize,
		OnProgress: func(uploaded, total int64) {
			log.Entry().Debugf("\rUploading: %d%% (%d/%d bytes)", (uploaded*100)/total, uploaded, total)
		},
	}
}
func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Uploaded += int64(n)
	if pw.OnProgress != nil {
		pw.OnProgress(pw.Uploaded, pw.Total)
	}
	return n, nil
}
