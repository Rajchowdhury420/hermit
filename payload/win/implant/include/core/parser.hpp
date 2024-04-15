#ifndef HERMIT_CORE_PARSER_HPP
#define HERMIT_CORE_PARSER_HPP

#include <windows.h>
#include <string>
#include <vector>

#include "core/crypt.hpp"
#include "core/json.hpp"
#include "core/stdout.hpp"
#include "core/utils.hpp"

using json = nlohmann::json;

namespace Parser
{
    json ParseTask(
        const std::vector<BYTE> task,
        BCRYPT_KEY_HANDLE hKey,
        std::vector<BYTE> iv
    );
}

#endif // HERMIT_CORE_PARSER_HPP
