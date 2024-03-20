#include "core/task.hpp"

namespace Task
{
    std::wstring Ls(const std::wstring& wDir)
    {
        std::wstring result;

        DWORD dwRet;
        WIN32_FIND_DATAW ffd;
        LARGE_INTEGER filesize;
        std::wstring wFilesize;
        WCHAR wTargetDir[MAX_PATH];
        size_t dirLength;
        WCHAR wBuffer[MAX_PATH];
        WCHAR** lppPart = {NULL};
        HANDLE hFind = INVALID_HANDLE_VALUE;

        StringCchLengthW(wDir.c_str(), MAX_PATH, &dirLength);
        if (dirLength > MAX_PATH)
        {
            return L"Error: Directory path is too long.";
        }

        StringCchCopyW(wTargetDir, MAX_PATH, wDir.c_str());
        StringCchCatW(wTargetDir, MAX_PATH, L"\\*");

        // Find the first file in the directory.
        hFind = FindFirstFile(wTargetDir, &ffd);
        if (hFind == INVALID_HANDLE_VALUE)
        {
            return L"Error: Could not find the first file in the directory.";
        }

        // Get the directory (absolute) path
        dwRet = GetFullPathNameW(
            wTargetDir,
            MAX_PATH,
            wBuffer,
            lppPart
        );
        if (dwRet == 0)
        {
            return L"Error: Could not get current directory.";
        }
        std::wstring wDirPath = std::wstring(wBuffer);
        result += std::wstring(L"Directory: ");
        result += std::wstring(wDirPath);
        result += std::wstring(L"\n\n");
        
        // List all files in the directory
        do
        {
            if (ffd.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY)
            {
                result += std::wstring(L"<D> ");
                result += std::wstring(ffd.cFileName);
                result += std::wstring(L"\n");
            }
            else
            {
                filesize.LowPart = ffd.nFileSizeLow;
                filesize.HighPart = ffd.nFileSizeHigh;
                wFilesize = std::to_wstring(filesize.QuadPart);

                result += std::wstring(L"<F> ");
                result += std::wstring(ffd.cFileName);
                result += std::wstring(L", ");
                result += wFilesize;
                result += std::wstring(L" bytes\n");
            }
        } while (FindNextFileW(hFind, &ffd) != 0);

        FindClose(hFind);
        return result;
    }
}