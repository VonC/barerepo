@echo off
setlocal

set GO111MODULE=on
set GOPROXY=https://proxy.golang.org

for %%i in ("%~dp0.") do SET "script_dir=%%~fi"
cd "%script_dir%" || echo "unable to cd to '%script_dir%'"&& exit /b 1
call  "%script_dir%\echos_macros.bat"
setlocal enabledelayedexpansion
for %%i in ("%~dp0.") do SET "dirname=%%~ni"

if exist "%script_dir%\senv.bat" ( call "%script_dir%\senv.bat" )

set "publish=TRUE"
call "%script_dir%\build.bat" amd
if errorlevel 1 (
    %_fatal% "%dirname% build.bat issue" 15
)

rem @echo on
for /f %%i in ('type version\version.txt') do set "appver=%%i"
%_task% "Must push %dirname%_%appver% to Nexus"


if not exist "%dirname%_%appver%" (
    fatal "No Linux file present named '%dirname%_%appver%'" 18
)

for /f %%i in ('type _creds.txt') do set "creds=%%i"
if "%creds%" == "" ( set "creds=unknown" )
if "%creds::=%" == "%creds%" (
    %_fatal% "Must provides Nexus upload creds ('%creds%') from private file '_creds.txt' as username:password" 16
)
if "%url%" == "" (
    %_fatal% "Must provides Nexus URL https://your.nexus/repostory" 18
)

if "%g%" == "" (
    %_fatal% "Must provides Nexus artifact group name 'xxx.yyy.zzz'" 17
)
set "a=%dirname%"
set "p=%g:.=/%"
set "v=%appver%"
set "f=%dirname%_%appver%"

set "repository=releases"
if not "%appver:-SNAPSHOT=%" == "%appver%" ( set "repository=snapshots")
set "url=%url%/%repository%/%p%/%a%/%v%"

call :curl_upload %f%

FOR /F "tokens=1 delims= " %%i in ('md5sum %f%') do ( set md5=%%i )
set "md5=%md5: =%"
%_info% "'%f%' md5='%md5%'"
echo %md5%>"%f%.md5"
call :curl_upload %f%.md5

FOR /F "tokens=1 delims= " %%i in ('sha1sum %f%') do ( set sha1=%%i )
set "sha1=%sha1: =%"
%_info% "'%f%' sha1='%md5%'"
echo %sha1%>"%f%.sha1"
call :curl_upload %f%.sha1

goto:eof

:curl_upload
set "ff=%1"
echo curl -v -k -u "%creds%" --upload-file "%ff%" %url%/%ff%
curl -v -k -u "%creds%" --upload-file "%ff%" %url%/%ff% 2>&1 | grep "201 Created"
if errorlevel 1 (
    %_fatal% "Unable to upload file '%ff%' to Nexus URL '%url%'" 17
)

goto:eof