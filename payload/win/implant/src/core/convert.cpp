#include "convert.hpp"

std::string VecCharToString(std::vector<char> chars)
{
    std::string s(chars.begin(), chars.end());
    return s;
}

std::string UTF8Encode(const std::wstring& wstr)
{
    if( wstr.empty() ) {
        return std::string();
    }

    int size_needed = WideCharToMultiByte(
        CP_UTF8,
        0,
        &wstr[0],
        (int)wstr.size(),
        NULL,
        0,
        NULL,
        NULL
    );

    std::string strTo( size_needed, 0 );

    WideCharToMultiByte(
        CP_UTF8,
        0,
        &wstr[0],
        (int)wstr.size(),
        &strTo[0],
        size_needed,
        NULL,
        NULL
    );

    return strTo;
}

std::wstring UTF8Decode(const std::string& str)
{
    if( str.empty() ) 
    {
        return std::wstring();
    }

    int size_needed = MultiByteToWideChar(
        CP_UTF8,
        0,
        &str[0],
        (int)str.size(),
        NULL,
        0
    );

    std::wstring wstrTo(size_needed, 0);

    MultiByteToWideChar(
        CP_UTF8,
        0,
        &str[0],
        (int)str.size(),
        &wstrTo[0],
        size_needed
    );

    return wstrTo;
}

// LPSTR (UTF-8) -> wchar_t* (UTF-16)
wchar_t* ConvertLPSTRToWCHAR_T(LPSTR lpStr)
{
    int wchars_num = MultiByteToWideChar(CP_UTF8, 0, lpStr, -1, NULL, 0);
    wchar_t* wstr = new wchar_t[wchars_num];
    MultiByteToWideChar(CP_UTF8, 0, lpStr, -1, wstr, wchars_num);
    return wstr;
}

// LPCWSTR (UTF-16) -> string (UTF-8)
std::string ConvertLPCWSTRToString(LPCWSTR lpcwStr)
{
    INT strLength = WideCharToMultiByte(CP_UTF8, 0, lpcwStr, -1, NULL, 0, NULL, NULL);
    std::string strData(strLength, 0);
    WideCharToMultiByte(CP_UTF8, 0, lpcwStr, -1, &strData[0], strLength, NULL, NULL);
    return strData;
}

// string -> LPCWSTR
LPCWSTR ConvertStringToLPCWSTR(const std::string& text)
{
    std::wstring wText = std::wstring(text.begin(), text.end());
    return wText.c_str();
}

