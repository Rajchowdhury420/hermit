package rpc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hideckies/hermit/pkg/common/meta"
	"github.com/hideckies/hermit/pkg/common/stdout"
	"github.com/hideckies/hermit/pkg/protobuf/commonpb"
	"github.com/hideckies/hermit/pkg/protobuf/rpcpb"
	"github.com/hideckies/hermit/pkg/server/handler"
	"github.com/hideckies/hermit/pkg/server/listener"
	"github.com/hideckies/hermit/pkg/server/loot"
	"github.com/hideckies/hermit/pkg/server/operator"
	"github.com/hideckies/hermit/pkg/server/payload"
	"github.com/hideckies/hermit/pkg/server/state"
	"github.com/hideckies/hermit/pkg/server/task"
)

type HermitRPCServer struct {
	rpcpb.UnimplementedHermitRPCServer
	serverState *state.ServerState
}

func (s *HermitRPCServer) SayHello(ctx context.Context, empty *commonpb.Empty) (*commonpb.Message, error) {
	return &commonpb.Message{Text: "Hello from Hermit"}, nil
}

func (s *HermitRPCServer) GetVersion(ctx context.Context, empty *commonpb.Empty) (*commonpb.Message, error) {
	return &commonpb.Message{Text: meta.GetVersion()}, nil
}

func (s *HermitRPCServer) OperatorRegister(
	ctx context.Context,
	ope *rpcpb.Operator,
) (*commonpb.Message, error) {
	newOpe := operator.NewOperator(0, ope.Uuid, ope.Name)
	err := s.serverState.DB.OperatorAdd(newOpe)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "You've been registered on the server successfully."}, nil
}

func (s *HermitRPCServer) OperatorDeleteByUuid(
	ctx context.Context,
	operatorUuid *commonpb.Uuid,
) (*commonpb.Message, error) {
	err := s.serverState.DB.OperatorDeleteByUuid(operatorUuid.Value)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "The operator deleted successfully."}, nil
}

func (s *HermitRPCServer) OperatorGetById(
	ctx context.Context,
	operatorId *commonpb.Id,
) (*rpcpb.Operator, error) {
	op, err := s.serverState.DB.OperatorGetById(uint(operatorId.Value))
	if err != nil {
		return nil, err
	}
	return &rpcpb.Operator{Id: int64(op.Id), Uuid: op.Uuid, Name: op.Name}, nil
}

func (s *HermitRPCServer) OperatorGetAll(
	empty *commonpb.Empty,
	stream rpcpb.HermitRPC_OperatorGetAllServer,
) error {
	ops, err := s.serverState.DB.OperatorGetAll()
	if err != nil {
		return err
	}

	for _, op := range ops {
		o := &rpcpb.Operator{
			Id:   int64(op.Id),
			Uuid: op.Uuid,
			Name: op.Name,
		}
		if err := stream.Send(o); err != nil {
			return err
		}
	}

	return nil
}

func (s *HermitRPCServer) ListenerStart(
	ctx context.Context,
	lis *rpcpb.Listener,
) (*commonpb.Message, error) {
	newLis := listener.NewListener(
		uint(lis.Id),
		lis.Uuid,
		lis.Name,
		lis.Protocol,
		lis.Host,
		uint16(lis.Port),
		strings.Split(lis.Domains, ","),
		lis.Active,
	)
	go handler.ListenerStart(newLis, s.serverState)
	err := s.serverState.Job.WaitListenerStart(s.serverState.DB, newLis)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "Listener started."}, nil
}

func (s *HermitRPCServer) ListenerStartById(
	ctx context.Context,
	listenerId *commonpb.Id,
) (*commonpb.Message, error) {
	lis, err := s.serverState.DB.ListenerGetById(uint(listenerId.Value))
	if err != nil {
		return nil, err
	}
	if lis.Active {
		return nil, fmt.Errorf("the listener is already running")
	}

	go handler.ListenerStart(lis, s.serverState)
	err = s.serverState.Job.WaitListenerStart(s.serverState.DB, lis)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "Listener started."}, nil
}

func (s *HermitRPCServer) ListenerStopById(
	ctx context.Context,
	listenerId *commonpb.Id,
) (*commonpb.Message, error) {
	lis, err := s.serverState.DB.ListenerGetById(uint(listenerId.Value))
	if err != nil {
		return nil, err
	}
	if !lis.Active {
		return nil, fmt.Errorf("listener already stopped")
	}

	s.serverState.Job.ChReqListenerQuit <- lis.Uuid
	err = s.serverState.Job.WaitListenerStop(s.serverState.DB, lis)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "Listener stoped."}, nil
}

func (s *HermitRPCServer) ListenerDeleteById(
	ctx context.Context,
	listenerId *commonpb.Id,
) (*commonpb.Message, error) {
	lis, err := s.serverState.DB.ListenerGetById(uint(listenerId.Value))
	if err != nil {
		return nil, err
	}
	if lis.Active {
		return nil, fmt.Errorf("the listener is running. Stop it before deleting")
	}

	err = s.serverState.DB.ListenerDeleteById(uint(listenerId.Value))
	if err != nil {
		return nil, err
	}

	// Delete folder
	listenerDir, err := meta.GetListenerDir(lis.Name, false)
	if err != nil {
		return nil, err
	}
	err = os.RemoveAll(listenerDir)
	if err != nil {
		return nil, err
	}

	return &commonpb.Message{Text: "Listener deleted."}, nil
}

func (s *HermitRPCServer) ListenerGetById(
	ctx context.Context,
	listenerId *commonpb.Id,
) (*rpcpb.Listener, error) {
	lis, err := s.serverState.DB.ListenerGetById(uint(listenerId.Value))
	if err != nil {
		return nil, err
	}
	return &rpcpb.Listener{
		Id:       int64(lis.Id),
		Uuid:     lis.Uuid,
		Name:     lis.Name,
		Protocol: lis.Protocol,
		Host:     lis.Addr,
		Domains:  strings.Join(lis.Domains, ","),
		Port:     int32(lis.Port),
		Active:   lis.Active,
	}, nil
}

func (s *HermitRPCServer) ListenerGetAll(
	empty *commonpb.Empty,
	stream rpcpb.HermitRPC_ListenerGetAllServer,
) error {
	liss, err := s.serverState.DB.ListenerGetAll()
	if err != nil {
		return err
	}

	for _, lis := range liss {
		l := &rpcpb.Listener{
			Id:       int64(lis.Id),
			Uuid:     lis.Uuid,
			Name:     lis.Name,
			Protocol: lis.Protocol,
			Host:     lis.Addr,
			Domains:  strings.Join(lis.Domains, ","),
			Port:     int32(lis.Port),
			Active:   lis.Active,
		}
		if err := stream.Send(l); err != nil {
			return err
		}
	}

	return nil
}

func (s *HermitRPCServer) PayloadImplantGenerate(
	ctx context.Context,
	imp *rpcpb.PayloadImplant,
) (*commonpb.Binary, error) {
	newImp := payload.NewImplant(
		uint(imp.Id),
		imp.Uuid,
		imp.Name,
		imp.Os,
		imp.Arch,
		imp.Format,
		imp.Lprotocol,
		imp.Lhost,
		uint16(imp.Lport),
		imp.Type,
		uint(imp.Sleep),
		uint(imp.Jitter),
		uint(imp.KillDate),
	)
	data, _, err := newImp.Generate(s.serverState)
	if err != nil {
		return nil, err
	}
	return &commonpb.Binary{Data: data}, nil
}

func (s *HermitRPCServer) PayloadStagerGenerate(
	ctx context.Context,
	stg *rpcpb.PayloadStager,
) (*commonpb.Binary, error) {
	newStg := payload.NewStager(
		uint(stg.Id),
		stg.Uuid,
		stg.Name,
		stg.Os,
		stg.Arch,
		stg.Format,
		stg.Lprotocol,
		stg.Lhost,
		uint16(stg.Lport),
		stg.Type,
		stg.Technique,
		stg.Process,
	)
	data, _, err := newStg.Generate(s.serverState)
	if err != nil {
		return nil, err
	}
	return &commonpb.Binary{Data: data}, nil
}

func (s *HermitRPCServer) PayloadShellcodeGenerate(
	ctx context.Context,
	shc *rpcpb.PayloadShellcode,
) (*commonpb.Binary, error) {
	newShc := payload.NewShellcode(
		uint(shc.Id),
		shc.Uuid,
		shc.Name,
		shc.Os,
		shc.Arch,
		shc.Format,
		shc.Lprotocol,
		shc.Lhost,
		uint16(shc.Lport),
		shc.Type,
		shc.TypeArgs,
	)
	data, _, err := newShc.Generate(s.serverState)
	if err != nil {
		return nil, err
	}
	return &commonpb.Binary{Data: data}, nil
}

func (s *HermitRPCServer) AgentGetById(
	ctx context.Context,
	agentId *commonpb.Id,
) (*rpcpb.Agent, error) {
	ag, err := s.serverState.DB.AgentGetById(uint(agentId.Value))
	if err != nil {
		return nil, err
	}
	return &rpcpb.Agent{
		Id:           int64(ag.Id),
		Uuid:         ag.Uuid,
		Name:         ag.Name,
		Ip:           ag.Ip,
		Os:           ag.OS,
		Arch:         ag.Arch,
		Hostname:     ag.Hostname,
		ListenerName: ag.ListenerName,
		Sleep:        int64(ag.Sleep),
		Jitter:       int64(ag.Jitter),
		KillDate:     int64(ag.KillDate),
	}, nil
}

func (s *HermitRPCServer) AgentGetAll(
	empty *commonpb.Empty,
	stream rpcpb.HermitRPC_AgentGetAllServer,
) error {
	ags, err := s.serverState.DB.AgentGetAll()
	if err != nil {
		return err
	}

	for _, ag := range ags {
		a := &rpcpb.Agent{
			Id:           int64(ag.Id),
			Uuid:         ag.Uuid,
			Name:         ag.Name,
			Ip:           ag.Ip,
			Os:           ag.OS,
			Arch:         ag.Arch,
			Hostname:     ag.Hostname,
			ListenerName: ag.ListenerName,
			Sleep:        int64(ag.Sleep),
			Jitter:       int64(ag.Jitter),
			KillDate:     int64(ag.KillDate),
		}
		if err := stream.Send(a); err != nil {
			return err
		}
	}
	return nil
}

func (s *HermitRPCServer) TaskSet(
	ctx context.Context,
	msg *commonpb.Message,
) (*commonpb.Message, error) {
	t := msg.GetText()
	err := task.SetTask(t, s.serverState.AgentMode.Name)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "Task set successfully."}, nil
}

func (s *HermitRPCServer) TaskClean(
	ctx context.Context,
	msg *commonpb.Empty,
) (*commonpb.Message, error) {
	err := meta.DeleteAllTasks(s.serverState.AgentMode.Name, false)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "All tasks deleted successfully."}, nil
}

func (s *HermitRPCServer) TaskList(
	ctx context.Context,
	empty *commonpb.Empty,
) (*commonpb.Message, error) {
	tasks, err := meta.ReadTasks(s.serverState.AgentMode.Name, false)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		stdout.LogWarn("Tasks not set.")
		return nil, fmt.Errorf("task not set")
	}
	return &commonpb.Message{Text: strings.Join(tasks, "\n")}, nil
}

func (s *HermitRPCServer) LootGetAll(
	ctx context.Context,
	empty *commonpb.Empty,
) (*commonpb.Message, error) {
	allLoot, err := loot.GetAllLoot(s.serverState.AgentMode.Name)
	if err != nil {
		stdout.LogFailed(fmt.Sprint(err))
		return nil, err
	}
	return &commonpb.Message{Text: allLoot}, nil
}

func (s *HermitRPCServer) LootClean(
	ctx context.Context,
	empty *commonpb.Empty,
) (*commonpb.Message, error) {
	err := meta.DeleteAllTaskResults(s.serverState.AgentMode.Name, false)
	if err != nil {
		return nil, err
	}
	return &commonpb.Message{Text: "All loot deleted successfully."}, nil
}
