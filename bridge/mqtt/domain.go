package mqtt

const (
	topicEndpointReboot    = "reboot"
	topicEndpointRepublish = "republish"
	topicEndpointReset     = "reset"
)

type Domain struct {
	client       Client
	topicService TopicService
}

func NewDomain(client Client, topicService TopicService) *Domain {
	return &Domain{
		client:       client,
		topicService: topicService,
	}
}

func (s *Domain) Republish() {
	topic := s.topicService.GetTopic(topicEndpointRepublish)
	s.client.Publish(topic, 0, false, "")
}

func (s *Domain) Reset() {
	topic := s.topicService.GetTopic(topicEndpointReset)
	s.client.Publish(topic, 0, false, "")
}

func (s *Domain) Reboot() {
	topic := s.topicService.GetTopic(topicEndpointReboot)
	s.client.Publish(topic, 0, false, "")
}

func (s *Domain) Publish(topic string, payload interface{}) {
	fullTopic := s.topicService.GetTopic(topic)
	s.client.Publish(fullTopic, 0, false, payload)
}

func (s *Domain) Subscribe(topic string, callback MessageHandler) {
	fullTopic := s.topicService.GetTopic(topic)
	s.client.Subscribe(fullTopic, 0, callback)
}
