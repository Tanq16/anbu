package scaffold

type Action string

const (
	Created  Action = "created"
	Existing Action = "existing"
	Deleted  Action = "deleted"
)

type Event struct {
	Action   Action `json:"action"`
	Resource string `json:"resource"`
	ID       string `json:"id"`
}

type recorder struct {
	events []Event
}

func (r *recorder) notify(action Action, resource, id string) {
	r.events = append(r.events, Event{Action: action, Resource: resource, ID: id})
}

const (
	resVPC         = "vpc"
	resIGW         = "internet gateway"
	resSubnet      = "subnet"
	resRouteTable  = "route table"
	resSecGroup    = "security group"
	resKeyPair     = "key pair"
	resScaffoldKey = "scaffold key"
)

const (
	nameVPC        = "sharingan-vpc"
	nameIGW        = "sharingan-igw"
	nameSubnet     = "sharingan-subnet"
	nameRouteTable = "sharingan-rtb"
	nameSecGroup   = "sharingan-sg"
	nameKeyPair    = "sharingan-key"
)
