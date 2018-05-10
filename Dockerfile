FROM golang:1.9-alpine

ARG exec_file

ENV PATH="/caravela:"${PATH}

WORKDIR /caravela
COPY $exec_file /caravela

# Expose the Overlay Port to outside (to other system nodes)
EXPOSE 8000
# Expose the CARAVELA's API Port to outside
EXPOSE 8001

VOLUME $HOME/.caravela

ENTRYPOINT ["caravela.exe"]