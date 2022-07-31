rem @echo off
SETLOCAL EnableDelayedExpansion

rem rmdir /Q /S dist
if NOT "%ERRORLEVEL%" == "0" goto error

rem npm run build
if NOT "%ERRORLEVEL%" == "0" goto error

del *.go
if NOT "%ERRORLEVEL%" == "0" goto error

echo package webgui > filelist.go
echo var FilesData map[string][]byte >> filelist.go
echo func init() { >> filelist.go
echo FilesData = make(map[string][]byte) >> filelist.go

set /a x=0
for /f "delims=" %%G in ('dir /A:-D /B dist') do (
    echo [%%G]
    2goarray.exe data_!x! webgui < dist\%%G > file_!x!.go
    if NOT "%ERRORLEVEL%" == "0" goto error

    echo FilesData["%%G"] = data_!x!  >> filelist.go
    set /a x+=1
)
echo } >> filelist.go

echo [OK]
exit /B 0

:error
echo [ERROR]
exit /B 1
