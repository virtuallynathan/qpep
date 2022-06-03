rem @echo off

echo ********************************
echo **** QPEP INSTALLER BUILDER ****
echo ********************************
echo.

set /A BUILD64=0
set /A BUILD32=0

if "%1" EQU "--build32" (
    set /A BUILD32=1
)
if "%1" EQU "--build64" (
    set /A BUILD64=1
)
if "%2" EQU "--build32" (
    set /A BUILD32=1
)
if "%2" EQU "--build64" (
    set /A BUILD64=1
)

if %BUILD64% EQU 0 (
    if %BUILD32% EQU 0 (
        echo Usage: installer.bat [--build32] [--build64]
        echo.
        goto fail
    )
)
if %BUILD32% EQU 0 (
    if %BUILD64% EQU 0 (
        echo Usage: installer.bat [--build32] [--build64]
        echo.
        goto fail
    )
)

ECHO [Cleanup]
DEL /S /Q build 2> NUL
RMDIR build\x64 2> NUL
RMDIR build\x86 2> NUL
RMDIR build\installer 2> NUL
RMDIR build 2> NUL

echo OK

ECHO [Preparation]
MKDIR build\  2> NUL
if %ERRORLEVEL% GEQ 1 goto fail
MKDIR build\x64\ 2> NUL
if %ERRORLEVEL% GEQ 1 goto fail
MKDIR build\x86\ 2> NUL
if %ERRORLEVEL% GEQ 1 goto fail

go clean
if %ERRORLEVEL% GEQ 1 goto fail

echo OK

set GOOS=windows
if %BUILD64% NEQ 0 (
    set GOARCH=amd64
    go clean -cache

    ECHO [Copy dependencies x64]
    COPY windivert\x64\* build\x64\
    if %ERRORLEVEL% GEQ 1 goto fail
    echo OK

    ECHO [Build x64 server/client]
    go build -o build\x64\qpep.exe
    if %ERRORLEVEL% GEQ 1 goto fail

    echo OK

    ECHO [Build x64 tray icon]
    pushd qpep-tray
    if %ERRORLEVEL% GEQ 1 goto fail
    go build -ldflags -H=windowsgui -o ..\build\x64\qpep-tray.exe
    if %ERRORLEVEL% GEQ 1 goto fail
    popd

    echo OK
)

if %BUILD32% NEQ 0 (
    set GOARCH=386
    go clean -cache

    ECHO [Copy dependencies x86]
    COPY windivert\x86\* build\x86\
    if %ERRORLEVEL% GEQ 1 goto fail
    echo OK

    ECHO [Build x86 server/client]
    go build -x -o build\x86\qpep.exe
    if %ERRORLEVEL% GEQ 1 goto fail

    echo OK

    ECHO [Build x86 tray icon]
    pushd qpep-tray
    if %ERRORLEVEL% GEQ 1 goto fail
    go build -ldflags -H=windowsgui -o ..\build\x86\qpep-tray.exe
    if %ERRORLEVEL% GEQ 1 goto fail
    popd

    echo OK
)

echo [Build of installer]
msbuild installer\installer.sln
if %ERRORLEVEL% GEQ 1 goto fail

echo ********************************
echo **** RESULT: SUCCESS        ****
echo ********************************
exit /B 0

:fail
echo ********************************
echo **** RESULT: FAILURE        ****
echo ********************************
exit /B 1

