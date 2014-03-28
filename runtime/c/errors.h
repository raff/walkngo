#ifndef _GO_RUNTIME_ERRORS_H
#define _GO_RUNTIME_ERRORS_H

namespace errors {
    error New(message string) {
        return error(message);
    }
}
