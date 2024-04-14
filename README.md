# Les Copaings Bot

Bot for the private server Discord "Les Copaings"

## Features

- Levels & XP
- Roles management
- Purge command

### XP

Functions:
- $xp_message(x;y) = \frac{0.025 x^{1.25}}{y^{-0.5}}+1$ where $x$ is the length of the message and $y$ is the diversity of the 
message (number of different rune)
- $xp_vocal(x)=0.01 x^{1.3}+1$ where $x$ is the time spent in vocal
- $level(x)=0.2 x^{0.5}$ where $x$ is the xp
- $level^{-1}(x)=(5x)^2$ where $x$ is the level
- $lose(x,y)= 10^{-2+\ln(x/85)}x^2+\lfloor y/100 \rfloor$ where $x$ is the inactivity time (hour) and $y$ is the xp

## Technologies

- Go 1.22
- anhgelus/gokord
