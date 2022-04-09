@echo off
setlocal enabledelayedexpansion

for %%i in ("%~dp0.") do SET "script_dir=%%~fi"
cd "%script_dir%"
for %%i in ("%~dp0.") do SET "dirname=%%~ni"

set "_args=%*"
if "%1" == "rel" ( 
    set "barg=rel"
    shift
    rem https://stackoverflow.com/questions/9363080/how-to-make-shift-work-with-in-batch-files
    set "_args=!_args:*%1 =!"
)

if "%1" == "amd" ( 
    set "barg=%barg% amd"
    shift
    set "_args=!_args:*%1 =!"
)
call build.bat %barg%
if errorlevel 1 (
    echo ERROR BUILD 1>&2
    exit /b 1
)
call run.bat %_args%