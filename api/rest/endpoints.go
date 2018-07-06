package rest

const ConfigurationBaseEndpoint = "/configuration"

const DiscoveryBaseEndpoint = "/discovery"
const DiscoveryOfferBaseEndpoint = DiscoveryBaseEndpoint + "/offer"
const DiscoveryNeighborOfferBaseEndpoint = DiscoveryBaseEndpoint + "/neighbor/offer"

const SchedulerBaseEndpoint = "/scheduler"
const SchedulerContainerBaseEndpoint = SchedulerBaseEndpoint + "/container"

const UserBaseEndpoint = "/user"
const UserContainerBaseEndpoint = UserBaseEndpoint + "/container"
const UserExitEndpoint = UserBaseEndpoint + "/exit"
