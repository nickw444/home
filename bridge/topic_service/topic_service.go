package topic_service

// TopicService is an interface for building topic strings
type TopicService interface {
	GetTopic(endpoint string) string
}

// PrefixedIDTopicService is a topic service implementation that returns a topic
// string with a prefix and ID.
type PrefixedIDTopicService struct {
	prefix     string
	identifier string
}

// NewPrefixedIDTopicService returns a PrefixedIDTopicService
func NewPrefixedIDTopicService(prefix string, identifier string) *PrefixedIDTopicService {
	return &PrefixedIDTopicService{
		identifier: identifier,
		prefix:     prefix,
	}
}

func (p *PrefixedIDTopicService) GetTopic(endpoint string) string {
	return p.prefix + "/" + p.identifier + "/" + endpoint
}
