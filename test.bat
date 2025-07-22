      
@ECHO OFF
:start
ECHO %TIME%
ping -n 2 -w 1000 localhost > nul
GOTO start
