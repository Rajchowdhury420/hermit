#ifndef HERMIT_TYPES_HPP
#define HERMIT_TYPES_HPP

#include <windows.h>
#include <winhttp.h>

// #pragma comment(lib, "winhttp.lib")

// A function to convert from string to wide string
#define WIDEN(x) WIDEN2(x)
#define WIDEN2(x) L##x

#ifdef PAYLOAD_TYPE
#define PAYLOAD_TYPE_W WIDEN(PAYLOAD_TYPE)
#endif

#ifdef PAYLOAD_TECHNIQUE
#define PAYLOAD_TECHNIQUE_W WIDEN(PAYLOAD_TECHNIQUE)
#endif

#ifdef PAYLOAD_PROCESS
#define PAYLOAD_PROCESS_W WIDEN(PAYLOAD_PROCESS)
#endif

//
#ifdef LISTENER_HOST
#define LISTENER_HOST_W WIDEN(LISTENER_HOST)
#endif

#ifdef LISTENER_USER_AGENT
#define LISTENER_USER_AGENT_W WIDEN(LISTENER_USER_AGENT)
#endif

#ifdef REQUEST_PATH_DOWNLOAD
#define REQUEST_PATH_DOWNLOAD_W WIDEN(REQUEST_PATH_DOWNLOAD)
#endif

#endif // HERMIT_TYPES_HPP