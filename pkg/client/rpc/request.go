package rpc

import (
	"context"
	"io"
	"strings"

	"github.com/hideckies/hermit/pkg/common/config"
	"github.com/hideckies/hermit/pkg/protobuf/commonpb"
	"github.com/hideckies/hermit/pkg/protobuf/rpcpb"
	"github.com/hideckies/hermit/pkg/server/agent"
	"github.com/hideckies/hermit/pkg/server/listener"
	"github.com/hideckies/hermit/pkg/server/operator"
	"github.com/hideckies/hermit/pkg/server/payload"
)

func RequestSayHello(c rpcpb.HermitRPCClient, ctx context.Context) (string, error) {
	r, err := c.SayHello(ctx, &commonpb.Empty{})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestGetVersion(c rpcpb.HermitRPCClient, ctx context.Context) (string, error) {
	r, err := c.GetVersion(ctx, &commonpb.Empty{})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestOperatorRegister(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	conf config.ClientConfig,
) (string, error) {
	r, err := c.OperatorRegister(ctx, &rpcpb.Operator{
		Id:   -1, // this value is not used
		Uuid: conf.Uuid,
		Name: conf.Operator,
	})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestOperatorDeleteByUuid(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	operatorUuid string,
) (string, error) {
	r, err := c.OperatorDeleteByUuid(ctx, &commonpb.Uuid{Value: operatorUuid})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestOperatorGetById(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	operatorId uint,
) (*operator.Operator, error) {
	r, err := c.OperatorGetById(ctx, &commonpb.Id{Value: int64(operatorId)})
	if err != nil {
		return nil, err
	}

	return operator.NewOperator(uint(r.GetId()), r.GetUuid(), r.GetName()), nil
}

func RequestOperatorGetAll(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
) ([]*operator.Operator, error) {
	stream, err := c.OperatorGetAll(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, err
	}

	ops := []*operator.Operator{}

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		newOp := operator.NewOperator(uint(data.GetId()), data.GetUuid(), data.GetName())
		ops = append(ops, newOp)
	}

	return ops, nil
}

func RequestListenerStart(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	lis *listener.Listener,
) (string, error) {
	r, err := c.ListenerStart(ctx, &rpcpb.Listener{
		Protocol: lis.Protocol,
		Host:     lis.Addr,
		Port:     int32(lis.Port),
		Domains:  strings.Join(lis.Domains, ","),
	})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestListenerStartById(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	listenerId uint,
) (string, error) {
	r, err := c.ListenerStartById(ctx, &commonpb.Id{Value: int64(listenerId)})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestListenerStopById(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	listenerId uint,
) (string, error) {
	r, err := c.ListenerStopById(ctx, &commonpb.Id{Value: int64(listenerId)})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestListenerDeleteById(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	listenerId uint,
) (string, error) {
	r, err := c.ListenerDeleteById(ctx, &commonpb.Id{Value: int64(listenerId)})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestListenerGetById(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	listenerId uint,
) (*listener.Listener, error) {
	r, err := c.ListenerGetById(ctx, &commonpb.Id{Value: int64(listenerId)})
	if err != nil {
		return nil, err
	}
	return listener.NewListener(
		uint(r.GetId()),
		r.GetUuid(),
		r.GetName(),
		r.GetProtocol(),
		r.GetHost(),
		uint16(r.GetPort()),
		strings.Split(r.GetDomains(), ","),
		r.GetActive(),
	), nil
}

func RequestListenerGetAll(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
) ([]*listener.Listener, error) {
	stream, err := c.ListenerGetAll(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, err
	}

	listeners := []*listener.Listener{}

	for {
		lis, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		newListener := listener.NewListener(
			uint(lis.Id),
			lis.Uuid,
			lis.Name,
			lis.Protocol,
			lis.Host,
			uint16(lis.Port),
			strings.Split(lis.Domains, ","),
			lis.Active,
		)
		listeners = append(listeners, newListener)
	}

	return listeners, nil
}

func RequestPayloadImplantGenerate(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	imp *payload.Implant,
) ([]byte, error) {
	r, err := c.PayloadImplantGenerate(ctx, &rpcpb.PayloadImplant{
		Os:       imp.Os,
		Arch:     imp.Arch,
		Format:   imp.Format,
		Lhost:    imp.Lhost,
		Lport:    int32(imp.Lport),
		Type:     imp.Type,
		Sleep:    int64(imp.Sleep),
		Jitter:   int64(imp.Jitter),
		KillDate: int64(imp.KillDate),
	})
	if err != nil {
		return []byte{}, err
	}
	return r.GetData(), nil
}

func RequestPayloadStagerGenerate(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	stg *payload.Stager,
) ([]byte, error) {
	r, err := c.PayloadStagerGenerate(ctx, &rpcpb.PayloadStager{
		Os:        stg.Os,
		Arch:      stg.Arch,
		Format:    stg.Format,
		Lhost:     stg.Lhost,
		Lport:     int32(stg.Lport),
		Type:      stg.Type,
		Technique: stg.Technique,
		Process:   stg.Process,
	})
	if err != nil {
		return []byte{}, err
	}
	return r.GetData(), nil
}

func RequestPayloadShellcodeGenerate(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	shc *payload.Shellcode,
) ([]byte, error) {
	r, err := c.PayloadShellcodeGenerate(ctx, &rpcpb.PayloadShellcode{
		Os:       shc.Os,
		Arch:     shc.Arch,
		Format:   shc.Format,
		Lhost:    shc.Lhost,
		Lport:    int32(shc.Lport),
		Type:     shc.Type,
		TypeArgs: shc.TypeArgs,
	})
	if err != nil {
		return []byte{}, err
	}
	return r.GetData(), nil
}

func RequestAgentGetById(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	agentId uint,
) (*agent.Agent, error) {
	r, err := c.AgentGetById(ctx, &commonpb.Id{Value: int64(agentId)})
	if err != nil {
		return nil, err
	}
	return agent.NewAgent(
		uint(r.GetId()),
		r.GetUuid(),
		r.GetName(),
		r.GetIp(),
		r.GetOs(),
		r.GetArch(),
		r.GetHostname(),
		r.GetListenerName(),
		uint(r.GetSleep()),
		uint(r.GetJitter()),
		uint(r.GetKillDate()),
	), nil
}

func RequestAgentGetAll(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
) ([]*agent.Agent, error) {
	stream, err := c.AgentGetAll(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, err
	}

	agents := []*agent.Agent{}

	for {
		ag, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		newAgent := agent.NewAgent(
			uint(ag.Id),
			ag.Uuid,
			ag.Name,
			ag.Ip,
			ag.Os,
			ag.Arch,
			ag.Hostname,
			ag.ListenerName,
			uint(ag.Sleep),
			uint(ag.Jitter),
			uint(ag.KillDate),
		)
		agents = append(agents, newAgent)
	}

	return agents, nil
}

func RequestTaskSet(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
	task string,
) error {
	_, err := c.TaskSet(ctx, &commonpb.Message{Text: task})
	if err != nil {
		return err
	}
	return nil
}

func RequestTaskClean(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
) error {
	_, err := c.TaskClean(ctx, &commonpb.Empty{})
	if err != nil {
		return err
	}
	return nil
}

func RequestTaskList(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
) (string, error) {
	r, err := c.TaskList(ctx, &commonpb.Empty{})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}

func RequestLootGetAll(
	c rpcpb.HermitRPCClient,
	ctx context.Context,
) (string, error) {
	r, err := c.LootGetAll(ctx, &commonpb.Empty{})
	if err != nil {
		return "", err
	}
	return r.GetText(), nil
}
