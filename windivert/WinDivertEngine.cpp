
extern "C" {
    #include "windivert_wrapper.h"
    #include "stdio.h"
}

HANDLE diverterHandle = INVALID_HANDLE_VALUE;

int InitializeWinDivertEngine() {
    if( diverterHandle != INVALID_HANDLE_VALUE )
        return DIVERT_ERROR_ALREADY_INIT;

    char filter[256] = "";
    snprintf(filter, 256, FILTERFMT, "192.168.1.100", 9090);

    printf(filter);
    printf("\n");

    diverterHandle = WinDivertOpen( filter, WINDIVERT_LAYER_NETWORK, WINDIVERT_PRIORITY_HIGHEST, 0 );
    if (diverterHandle == INVALID_HANDLE_VALUE)
        return DIVERT_ERROR_NOTINITILIZED;

    return DIVERT_OK;
}

int CloseWinDivertEngine() {
    if( diverterHandle == INVALID_HANDLE_VALUE )
        return DIVERT_ERROR_NOTINITILIZED;

    if( WinDivertClose(diverterHandle) == TRUE )
        return DIVERT_ERROR_FAILED;

    diverterHandle = INVALID_HANDLE_VALUE;
    return DIVERT_OK;
}
