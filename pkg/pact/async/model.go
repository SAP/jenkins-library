package async

// AsyncAPISpec Represents an AsyncAPISpec json file.
type AsyncAPISpec struct {
	AsyncAPI   string             `json:"asyncapi"`
	Channels   map[string]Channel `json:"channels"`
	Components Components         `json:"components"`
}

// Components represents the outer wrapper object in the json file for components.
type Components struct {
	Schemas map[string]Schema `json:"schemas"`
}

// Schema represents a component schema.
type Schema struct {
	Type       string
	Properties map[string]Property    `json:"properties"`
	Examples   map[string]interface{} `json:"example"`
}

// Property represents a component schema property.
type Property struct {
	Type   string   `json:"type"`
	Ref    string   `json:"$ref"`
	Items  Items    `json:"items"`
	Enum   []string `json:"enum"`
	Format string   `json:"format"`
}

type Items struct {
	Ref  string `json:"$ref"`
	Type string `json:"type"`
}

// Channel represents the channels map in AsyncAPI Spec.
type Channel struct {
	Publish struct {
		Bindings struct {
			Kafka struct{} `json:"kafka"`
		} `json:"bindings"`
		Message struct {
			Name    string `json:"name"`
			Title   string `json:"title"`
			Payload struct {
				Ref string `json:"$ref"`
			} `json:"payload"`
		} `json:"message"`
	} `json:"publish"`
}


// AsyncPactSpec Represents an AsyncPactSpec json file
type AsyncPactSpec struct {
	Consumer Consumer  `json:"consumer"`
	Provider Provider  `json:"provider"`
	Messages []Message `json:"messages"`
}

// Message represents the outer wrapper object in the json file for messages.
type Message struct {
	ID            string                 `json:"_id"`
	Description   string                 `json:"description"`
	MetaData      Meta                   `json:"metaData"`
	Contents      map[string]interface{} `json:"contents"`
	MatchingRules *MatchingRules         `json:"matchingRules"`
}

// Consumer represents the consumer of the given contract
type Consumer struct {
	Name string `json:"name"`
}

// Provider represents the provider of the given contract
type Provider struct {
	Name string `json:"name"`
}

// Meta represents the metaData in the AsyncPactSpec
type Meta struct {
	Topic       string `json:"Topic"`
	ContentType string `json:"contentType"`
}

type MatchingRules struct {
	Body map[string]MatchingRule `json:"body"`
}

type MatchingRule struct {
	Matchers []Matcher `json:"matchers"`
	Combine  string    `json:"combine"`
}

type Matcher struct {
	Match string `json:"match"`
	Date  string `json:"date"`
}
