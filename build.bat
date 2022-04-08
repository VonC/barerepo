@echo off
setlocal

set GO111MODULE=on
set GOPROXY=https://proxy.golang.org

for %%i in ("%~dp0.") do SET "script_dir=%%~fi"
cd "%script_dir%" || echo "unable to cd to '%script_dir%'"&& exit /b 1
call  "%script_dir%\echos_macros.bat"
setlocal enabledelayedexpansion
for %%i in ("%~dp0.") do SET "dirname=%%~ni"

if not exist "%script_dir%\senv.bat" ( copy "%script_dir%\senv.bat.tpl" "%script_dir%\senv.bat" )
if exist "%script_dir%\senv.bat" ( call "%script_dir%\senv.bat" )

if "%1" == "rel" ( 
    sed -i "s/-SNAPSHOT//g" "version\version.txt"
    shift
)
for /f %%i in ('type version\version.txt') do set "appver=%%i"
echo "Application Version: '%appver%'"

rem https://medium.com/@joshroppo/setting-go-1-5-variables-at-compile-time-for-versioning-5b30a965d33e
for /f %%i in ('git describe --long --tags --dirty --always') do set gitver=%%i
for /f %%i in ('git describe --tags 2^>NUL') do set VERSION=%%i
if "%VERSION%" == "" ( set "VERSION=%appver%")
echo VERSION='%VERSION%'
set "snap=FALSE"
set "askForNewSnapshot=FALSE"
if not "%VERSION:-=%" == "%VERSION%" (
    set "snap=TRUE-SNAP"
    set "askForNewSnapshot=new commits"
)
if not "%gitver:-dirty=%" == "%gitver%" (
    set "snap=!snap!-dirty"
    if "%askForNewSnapshot%" == "FALSE" (
        set "askForNewSnapshot=dirty"
    ) else (
        set "askForNewSnapshot=%askForNewSnapshot%, dirty"
    )
)
echo snap detection '%snap%'
if not "%snap%" == "FALSE" (
    if not "%snap:-SNAP=%" == "%snap%" (
        set "todelete=-%VERSION:*-=%"
        call set "VERSION=%%VERSION:!todelete!=%%"
    )
    echo snap !VERSION! with todelete='!todelete!', askForNewSnapshot='%askForNewSnapshot%'
) else (
    echo release %VERSION%
)
echo "Git VERSION='%VERSION%'"

set "makeNewRelease=FALSE"
if not "v%appver%" == "%VERSION%" (
    if not "%appver:-SNAP=%" == "%appver%" (
        %_ok% "Building '%appver%' from last release Git tag '%VERSION%' (snapshot)"
    ) else (
        set "makeNewRelease=TRUE"
    )
)

if "%makeNewRelease%" == "TRUE" (
    %_warning% "New release detected '%appver%', differs from last release Git tag '%VERSION%'"
    %_task% "Must commit and tag new v%appver%."
    git add .
    if errorlevel 1 ( %_fatal% "ERROR unable to add before tagging '%appver%'" 40)
    git commit -m "New release '%appver%'"
    if errorlevel 1 ( %_fatal% "ERROR unable to commit before tagging '%appver%'" 41)
    git tag -m "v%appver%" v%appver%
    if errorlevel 1 ( %_fatal% "ERROR unable to tag 'v%appver%'" 42)
    set VERSION="v%appver%"
    for /f %%i in ('git describe --long --tags --dirty --always') do set gitver=%%i
    set "snap=FALSE"
    set "todelete="
)

if "v%appver%" == "%VERSION%" (
    if not "%askForNewSnapshot%" == "FALSE" (
        %_warning% "New modifications detected since last release '%VERSION%' (%askForNewSnapshot%)"
        git diff --cached --quiet
        if errorlevel 1 (
            %_fatal% "Please commit or reset your indexed/staged changes first, to allow version.txt modification and individual commit" 111
        )
        %_task% "Specify the new SNAPSHOT version to do"
        FOR /F "tokens=1,2,3 delims=." %%i in ("%appver%") do (
            set maj=%%i
            set min=%%j
            set fix=%%k
        )
        echo "Major='!maj!', Minor='!min!', Fix='!fix!'"
        set nfix=!fix!
        set /A nfix+=1
        ECHO 1. Fix   update: !maj!.!min!.!nfix!-SNAPSHOT
        set nmin=!min!
        set /A nmin+=1
        ECHO 2. Minor update: !maj!.!nmin!.0-SNAPSHOT
        set nmaj=!maj!
        set /A nmaj+=1
        ECHO 3. Major update: !nmaj!.0.0-SNAPSHOT
        choice /C 123 /M "Select the new snapshot version you want to make next"
        set c=!errorlevel!
        echo "Choice '!c!'"

        if "!c!" == "1" ( set "appver=!maj!.!min!.!nfix!-SNAPSHOT" )
        if "!c!" == "2" ( set "appver=!maj!.!nmin!.0-SNAPSHOT" )
        if "!c!" == "3" ( set "appver=!nmaj!.0.0-SNAPSHOT" )
        echo !appver!>"version\version.txt"
        git add "version\version.txt"
        if errorlevel 1 ( %_fatal% "ERROR unable to add version\version.txt" 112 )
        git commit -m "Begin new '!appver!' from previous release '%VERSION%'"
        if errorlevel 1 ( %_fatal% "ERROR unable to commit version\version.txt" 112 )
    )
)

rem https://superuser.com/questions/1287756/how-can-i-get-the-date-in-a-locale-independent-format-in-a-batch-file
rem https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/get-date?view=powershell-7.1
rem C:\Windows\System32\WindowsPowershell\v1.0\powershell -Command "Get-Date -format 'yyyy-MM-dd_HH-mm-ss K'"
%+@% for /f %%a in ('powershell -Command "Get-Date -format yyyy-MM-dd_HH-mm-ss"') do set dtStamp=%%a
rem SET dtStamp
echo "dtStamp='%dtStamp%'"

set outputname=%dirname%.exe

if "%1" == "amd" (
    set GOARCH=amd64
    set GOOS=linux
    set "outputname=%dirname%_%appver%"
    %_info% "AMD build requested for %dirname%"
    set "fflag=-gcflags="all=-N -l" "
)

%_info% "Start Building"
go build %fflag%-ldflags "-X %dirname%/version.GitTag=%gitver% -X %dirname%/version.BuildUser=%USERNAME% -X %dirname%/version.Version=%VERSION% -X %dirname%/version.BuildDate=%dtStamp%" -o %outputname%

if errorlevel 1 (
    %_fatal% "ERROR BUILD %dirname%" 3
)
