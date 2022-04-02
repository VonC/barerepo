@echo off

REM set GOROOT=C:\path\to\Go
REM set g=xxx.yyy.zzz
REM set url=https://your.nexus/repository

doskey a=all.bat $*
doskey b=build.bat $*
doskey r=run.bat $*
doskey p=publish.bat $*

doskey arel=all.bat rel $*
doskey brel=build.bat rel $*

doskey av=all.bat -v $*
doskey avv=all.bat -vv $*
doskey avvv=all.bat -vvv $*
doskey ave=all.bat version $*
doskey avev=all.bat version -v $*
doskey avevv=all.bat version -vv $*

doskey rv=run.bat -v $*
doskey rvv=run.bat -vv $*
doskey rvvv=run.bat -vvv $*
doskey rve=run.bat version $*
doskey rvev=run.bat version -v $*
doskey rvevv=run.bat version -vv $*