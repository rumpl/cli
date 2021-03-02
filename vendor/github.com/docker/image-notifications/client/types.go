package client

import "time"

// ErrorJSON ...
type ErrorJSON struct {
	Message string `json:"message"`
}

// CVE holds information about security vulnerabilities of an image
type CVE struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// ImageInfo holds various information about a particular image
type ImageInfo struct {
	Digest         string     `json:"digest"`           // Digest is the image digest
	Tags           []string   `json:"tags"`             // Tags is the list of tags pointing to this digest
	Deprecated     bool       `json:"deprecated"`       // Deprecated is true if the image should be considered as deprecated
	EndOfLife      *time.Time `json:"end_of_life"`      // EndOfLife represents the date when the support for this image ends
	Changelog      *string    `json:"changelog"`        // Changelog is the link to the changelog
	ReleaseNotes   *string    `json:"release_notes"`    // ReleaseNotes is the link to the release note
	Size           int64      `json:"size"`             // Size is the size (in bytes) of the image
	LastUpdateDate *time.Time `json:"last_update_date"` // LastUpdateDate
	Repository     *string    `json:"repository"`       // Repoistory is the link to the repository of the image
	License        *string    `json:"license"`          // License
	Hub            *string    `json:"hub"`              // Hub
	CVEs           []CVE      `json:"cves"`             // CVEs is the list of fixed CVEs in this image
}

// ImageInfoResponse is the reponse for an image info
type ImageInfoResponse struct {
	ImageInfos []*ImageInfo `json:"image_infos"`
}
