package main

// ContentCollection holds items information
type ContentCollection struct {
	UUID             string `json:"uuid,omitempty"`
	Items            []Item `json:"items,omitempty"`
	PublishReference string `json:"publishReference,omitempty"`
	LastModified     string `json:"lastModified,omitempty"`
	CollectionType   string `json:"type,omitempty"`
}

// Item within content collection
type Item struct {
	UUID string `json:"uuid,omitempty"`
}

// MappedContent top level type
type MappedContent struct {
	Payload      ContentCollection `json:"payload,omitempty"`
	ContentURI   string            `json:"contentUri,omitempty"`
	LastModified string            `json:"lastModified,omitempty"`
	UUID         string            `json:"uuid,omitempty"`
}
