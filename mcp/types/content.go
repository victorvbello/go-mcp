package types

const (
	TEXT_CONTENT_TYPE              = "text"
	IMAGE_CONTENT_TYPE             = "image"
	AUDIO_CONTENT_TYPE             = "audio"
	EMBEDDED_RESOURCE_CONTENT_TYPE = "resource"
)

type Content interface {
	TypeOfContent() string
}

type BaseContent struct {
	Type string `json:"type"`
}

//Text provided to or from an LLM.
type TextContent struct {
	//type: "text"
	BaseContent
	//The text content of the message.
	Text string `json:"text"`
}

func (t *TextContent) TypeOfContent() string { return TEXT_CONTENT_TYPE }

func NewTextContent(txt string) *TextContent {
	ntxt := new(TextContent)
	ntxt.Type = "text"
	ntxt.Text = txt
	return ntxt
}

//An image provided to or from an LLM.
type ImageContent struct {
	//type: "image"
	BaseContent
	//The base64-encoded image data.
	Data string `json:"data"`
	//The MIME type of the image. Different providers may support different image types.
	MIMEType string `json:"mimeType"`
}

func (i *ImageContent) TypeOfContent() string { return IMAGE_CONTENT_TYPE }

func NewImageContent(Data, MIMEType string) *ImageContent {
	c := new(ImageContent)
	c.Type = "image"
	c.Data = Data
	c.MIMEType = MIMEType
	return c
}

//Audio provided to or from an LLM.
type AudioContent struct {
	//type: "audio"
	BaseContent
	//The base64-encoded audio data.
	Data string `json:"data"`
	//The MIME type of the audio. Different providers may support different audio types.
	MIMEType string `json:"mimeType"`
}

func (a *AudioContent) TypeOfContent() string { return AUDIO_CONTENT_TYPE }

func NewAudioContent(Data, MIMEType string) *AudioContent {
	c := new(AudioContent)
	c.Type = "audio"
	c.Data = Data
	c.MIMEType = MIMEType
	return c
}

//The contents of a resource, embedded into a prompt or tool call result.
//
//It is up to the client how best to render embedded resources for the benefit
//of the LLM and/or the user.
type EmbeddedResource struct {
	//type: "resource"
	BaseContent
	Resource ResourceContents `json:"resource"` //TextResourceContents/BlobResourceContents;
}

func (e *EmbeddedResource) TypeOfContent() string { return EMBEDDED_RESOURCE_CONTENT_TYPE }

func NewEmbeddedResource(Resource ResourceContents) *EmbeddedResource {
	c := new(EmbeddedResource)
	c.Type = "resource"
	c.Resource = Resource
	return c
}
