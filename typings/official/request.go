package official

type APIRequest struct {
	Messages  []apiMessage `json:"messages"`
	Stream    bool         `json:"stream"`
	Model     string       `json:"model"`
	PluginIDs []string     `json:"plugin_ids"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
