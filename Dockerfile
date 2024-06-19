FROM mcr.microsoft.com/windows/servercore:ltsc2022
USER ContainerAdministrator
RUN cd C:
RUN setx /M PATH "%PATH%;C:/"
COPY ./bins/client.exe ./
COPY ./bins/server.exe ./
COPY ./bins/nodeexporter.exe ./
EXPOSE 4444
EXPOSE 4445
EXPOSE 9100