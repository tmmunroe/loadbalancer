package loadbalancer

type PingService interface {
	Ping(args *PingArgs, reply *PingReply) error
}

type MathService interface {
	Add(args *MathArgs, reply *MathReply) error
	Multiply(args *MathArgs, reply *MathReply) error
}

type RegistrationService interface {
	Register(args *RegisterArgs, reply *RegisterReply) error
}
