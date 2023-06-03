package bgpd

// BGP is an instance of the BGP speaker.
// NOTE: the struct is incomplete. this struct only contains the used fields.
type BGP struct {
	// Routes is the sequence of the BGP routes.
	// this key is notated with the CIDR-block.
	Routes map[string][]BGPRoute `json:"routes"`
}

// BGPRoute is a route that contains various attributes in BGP.
// NOTE: the struct is incomplete. this struct only contains the used fields.
type BGPRoute struct {
}
