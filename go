#!/bin/bash
export SCRIPT_PID=$$
export BASE_DIR=`pwd`

# import external configuration file entries
. ./go.conf

export APP_UID=$(id -u)
export APP_GID=$(id -g)
export WARS_DIR="${BASE_DIR}/wars"
export CONF_DIR="${BASE_DIR}/conf"
export PROPS_DIR="${BASE_DIR}/properties"
export LOGS_DIR="${BASE_DIR}/logs"
export CERTS_DIR="${BASE_DIR}/certs"
export KEY_DIR="/home/dockrusr/.ssh"
export WEBAPPS_VOLUME_NAME="${INSTANCE_NAME}_webapps_volume"

usage(){
	echo "Usage:"
	echo " $0 <option>"
	echo "    start		- starts the container in the background"
	echo "    restart	- restarts the container"
	echo "    stop		- stops the container"
	echo "    copy-files	- copy required files"
	echo "    update-ports	- update the standard port numbers in the configs to your selected ports"
	echo "    shell		- creates a shell in the container to look around"
	echo "    build		- build the docker image that will be run (reads Dockerfile)"
	exit 0
}

create_required_directories(){
	if [ ! -d ${CONF_DIR} ]
	then
		echo "Creating local conf directory: ${CONF_DIR}"
		mkdir ${CONF_DIR}
	fi

#	if [ ! -d ${WARS_DIR} ]
#	then
#		echo "Creating local warfile directory: ${WARS_DIR}"
#		mkdir ${WARS_DIR}
#	fi

	if [ ! -d "${LOGS_DIR}" ]; then
		echo "Creating local logging directory: ${LOGS_DIR}"
      		mkdir "${LOGS_DIR}"
	fi


	if [ ! -d "${PROPS_DIR}" ]; then
		echo "Creating local properties directory: ${PROPS_DIR}"
      		mkdir "${PROPS_DIR}"
	fi

	if [ ! -d "${CERTS_DIR}" ]; then
		echo "Creating local certs directory: ${CERTS_DIR}"
      		mkdir "${CERTS_DIR}"
	fi

}

get_required_files(){
	if [ ${TOMCAT_INSTALL} = "true" ]; then

		SERVER_XML="${CONF_DIR}/server.xml"
		if [ ! -f "${SERVER_XML}" ]; then
			cp -av dist/server.xml ${SERVER_XML}
 		fi
	
		TOMCAT_USERS_XML="${CONF_DIR}/tomcat-users.xml"
		if [ ! -f ${TOMCAT_USERS_XML} ]
		then
			cp -av dist/tomcat-users.xml ${TOMCAT_USERS_XML}
			
		fi
	
		MANAGER_CONTEXT_XML="${CONF_DIR}/manager-context.xml"
		if [ ! -f ${MANAGER_CONTEXT_XML} ]
		then
			cp -av dist/manager-context.xml ${MANAGER_CONTEXT_XML}
		fi
	
		MANAGER_WEB_XML="${CONF_DIR}/manager-web.xml"
		if [ ! -f ${MANAGER_WEB_XML} ]
		then
			cp -av dist/manager-web.xml ${MANAGER_WEB_XML}
		fi
	
		MAVEN_SETTINGS_XML="${CONF_DIR}/maven-settings.xml"
			if [ ! -f ${MAVEN_SETTINGS_XML} ]
		then
			cp -av dist/maven-settings.xml ${MAVEN_SETTINGS_XML}
		fi
	
	fi
}

generate_ssl_key(){
	if [ ${TOMCAT_INSTALL} = "true" ]; then

		KEYSTORE_FILE="${CERTS_DIR}/keystore.jks"

		if [ ! -f ${KEYSTORE_FILE} ]
		then
			echo "No keystore found, creating a self-signed keystore just to get us up and running... YOU WILL NEED TO UPDATE WITH A REAL SSL KEY"
			docker run -ti \
			-v ${CERTS_DIR}:/vm-certs \
			${INSTANCE_NAME} \
			keytool -genkey \
			-keystore /vm-certs/keystore.jks \
			-keyalg RSA \
			-keysize 2048 \
			-validity 10000 \
			-alias app \
			-dname "cn='temp.itg.ti.com' o='Texas Instruments', ou='IT', st='Texas',  c='US'" \
			-storepass Ry9coo2J \
			-keypass Ry9coo2J 
		fi
	fi
}


generate_start_option_lines(){
(
        for MAPPING in `cat go_mappings.conf | grep -v ^#`
        do
                TYPE=`echo ${MAPPING} | awk -F \| '{print $1}'`

                if [ ${TYPE} = "BIND" ]; then
                        BIND=`echo ${MAPPING} | awk -F \| '{print $1}'`
                        HOST_LOCATION=`echo ${MAPPING} | awk -F \| '{print $3}'`
                        FULL_HOST_LOCATION=`readlink -f ${HOST_LOCATION}`
                        GUEST_LOCATION=`echo ${MAPPING} | awk -F \| '{print $4}'`
                        echo "-v ${FULL_HOST_LOCATION}:${GUEST_LOCATION} "

                fi

                if [ ${TYPE} = "VOLUME" ]; then
                        VOLUME=`echo ${MAPPING} | awk -F \| '{print $1}'`
                        VOLUME_NAME=`echo ${MAPPING} | awk -F \| '{print $2}'`
                        MOUNT_LOCATION=`echo ${MAPPING} | awk -F \| '{print $3}'`
                        echo "--mount source=${VOLUME_NAME},target=${MOUNT_LOCATION} "

                fi

                if [ ${TYPE} = "WARFILE" ]; then
                        HOST_LOCATION=`echo ${MAPPING} | awk -F \| '{print $2}'`
                        FULL_HOST_LOCATION=`readlink -f ${HOST_LOCATION}`
                        GUEST_LOCATION=`echo ${MAPPING} | awk -F \| '{print $3}'`
                        echo "-v ${FULL_HOST_LOCATION}:${GUEST_LOCATION} "
                fi

                if [ ${FROM} = "PORT" ]; then
                        HOST_PORT=`echo ${MAPPING} | awk -F \| '{print $2}'`
                        GUEST_PORT=`echo ${MAPPING} | awk -F \| '{print $3}'`
                        echo "-p ${HOST_PORT}:${GUEST_PORT} "
                fi
        done

	# WARFILES:
	
#	for FILE in `ls -1 ${WARS_DIR}`
#	do
#		FULL_HOST_LOCATION=`readlink -f ${FILE}`
#                echo "-v ${FULL_HOST_LOCATION}:/usr/local/tomcat/webapps/${FILE} "
#	done

        cd ${BASE_DIR}
)
}



validate_go_mapping_lines(){
(
        for MAPPING in `cat go_mappings.conf | grep -v ^#`
        do 
                TYPE=`echo ${MAPPING} | awk -F \| '{print $1}'`

                if [ ${TYPE} = "BIND" ]; then
			BIND=`echo ${MAPPING} | awk -F \| '{print $1}'`
			HOST_LOCATION=`echo ${MAPPING} | awk -F \| '{print $3}'`
			
			if [ ${BIND} = 'FILE' ]; then
				if [ ! -f ${HOST_LOCATION} ]; then
					echo "ERROR --- ${HOST_LOCATION} of type ${BIND} not found!, exiting..."	
					kill_script
				fi  
			elif [ ${BIND} = 'DIRECTORY' ]; then
				if [ ! -d ${HOST_LOCATION} ]; then
					echo "ERROR --- ${HOST_LOCATION} of type ${BIND} not found!, exiting..."	
					kill_script
				fi  
			fi		


                fi

                if [ ${TYPE} = "VOLUME" ]; then
                        VOLUME_NAME=`echo ${MAPPING} | awk -F \| '{print $2}'`
                        AUTO_CREATE_VOLUME=`echo ${MAPPING} | awk -F \| '{print $4}'`
			docker volume inspect ${VOLUME_NAME}
			if [ $? != 0 ]; then
				if [ ${AUTO_CREATE_VOLUME} = "TRUE" ]; then
					echo "Auto create option for volume: ${AUTO_CREATE_VOLUME} set to TRUE, creating..."
					docker volume create ${VOLUME_NAME}
					if [ $? != 0 ]; then
                                		echo "ERROR ---  Cannot auto-create volume listed in go_mappings.conf file: ${VOLUME_NAME} EXITING..."
                              			kill_script
					else
						echo "Volume: ${VOLUME_NAME} Created automatically"
					fi
				fi
				echo "ERROR ---  Cannot locate volume listed in go_mappings.conf file: ${VOLUME_NAME}"
				kill_script
			fi

                fi

                if [ ${TYPE} = "WARFILE" ]; then
			HOST_LOCATION=`echo ${MAPPING} | awk -F \| '{print $2}'`
			GUEST_LOCATION=`echo ${MAPPING} | awk -F \| '{print $3}'`
			if [ ! -f ${HOST_LOCATION} ]; then
				echo "ERROR ---  Cannot find war file listed in go_mappings.conf file: ${WARS_DIR}/${HOST_LOCATION}"
				kill_script
			fi

                fi

        done

        cd ${BASE_DIR}
)
}

#get_wars(){
#	if [ ${TOMCAT_INSTALL} = "true" ]; then
#		if [ ! -f "${WARS_DIR}/ffw-properties-manager.war" ]; then
#			echo "${WARS_DIR}/ffw-properties-manager.war does not exist... Getting it from artifactory..."
#			cd ${WARS_DIR} && \
#			curl -O "https://artifactory.itg.ti.com/artifactory/generic-tmg-prod-local/pmo/sa/docker-demo/ffw-properties-manager.war" && \
#
#			if [ $? -eq 0 ]
#			then
#				cd ${BASE_DIR}/
#			else
#				echo "ERROR:  could not retrieve ffw-properties-manager.war"
#				exit 44
#			fi
#		fi
#	fi
#}

change_ports(){
	util_scripts/change_ports
}

remove_git(){
	if [ -d .git ]; then
		if [ ${REMOVE_GIT} = "true" ]; then
			echo "Removing git directory to stop any changes here making it to bitbucket"
			rm -rf .git
		fi
	fi
}

update_volume_permissions(){
	echo "Updating volume permissions using INSTANCE_NAME: ${INSTANCE_NAME}"

	docker run \
	--user root \
	-ti \
	--mount source=${WEBAPPS_VOLUME_NAME},target=/tmp/v_mount \
	--rm \
	--name=tomcat-temp-delete ${INSTANCE_NAME} \
	chown -R ${IMAGE_USER}:${IMAGE_USER} /tmp/v_mount/
}

copy_tomcat_apps(){
	update_volume_permissions

	echo "Copying manager application to webapps directory using INSTANCE_NAME: ${INSTANCE_NAME}"
	docker run -ti \
	--mount source=${WEBAPPS_VOLUME_NAME},target=/tmp/v_mount \
	--rm \
	--name=tomcat-temp-delete ${INSTANCE_NAME} \
	cp -av /usr/local/tomcat/webapps.dist/manager /tmp/v_mount/

	echo "Copying ROOT application to webapps directory using INSTANCE_NAME: ${INSTANCE_NAME}"
	docker run -ti \
	--mount source=${WEBAPPS_VOLUME_NAME},target=/tmp/v_mount \
	--rm \
	--name=tomcat-temp-delete ${INSTANCE_NAME} \
	cp -av /usr/local/tomcat/webapps.dist/ROOT /tmp/v_mount/

	echo "Copying host-manager application to webapps directory using INSTANCE_NAME: ${INSTANCE_NAME}"
	docker run -ti \
	--mount source=${WEBAPPS_VOLUME_NAME},target=/tmp/v_mount \
	--rm \
	--name=tomcat-temp-delete ${INSTANCE_NAME} \
	cp -av /usr/local/tomcat/webapps.dist/host-manager /tmp/v_mount/
}

kill_script(){
	kill -s TERM ${SCRIPT_PID}
}

case "$1" in
  start)
   #get_wars
   validate_go_mapping_lines
   create_required_directories
   get_required_files
   generate_ssl_key
   remove_git
	

  	GENERATED_OPTIONS=$(generate_start_option_lines)

	echo "Starting with these options: ${GENERATED_OPTIONS}"

	if [ ${TESTING_GO_FILE} = "true" ]; then
		echo "Not starting instance due to TESTING_GO_FILE set to true"
	else
	docker run -d \
     	-e CATALINA_OPTS=\"${CATALINA_OPTS}\" \
      	-e JAVA_OPTS=\"${JAVA_OPTS}\" \
	--user ${IMAGE_USER}  \
	${GENERATED_OPTIONS} \
	--name ${INSTANCE_NAME} \
	--hostname ${INSTANCE_NAME} \
 	--restart=${DOCKER_RESTART_POLICY} \
   	${INSTANCE_NAME}
	fi
  ;;
  exec)
    docker exec -ti ${INSTANCE_NAME} ${@:2}
  ;;
  root)
    docker exec -u root -ti ${INSTANCE_NAME} bash 
  ;;
  logs)
    docker logs ${INSTANCE_NAME} 2>&1
  ;;
  tail)
    docker logs --follow -n1 ${INSTANCE_NAME} 2>&1
  ;;
  #simple stop/start
  restart)
    $0 stop
    $0 start
  ;;
  stop)
    docker stop ${INSTANCE_NAME};
    docker rm ${INSTANCE_NAME};
  ;;
  build)
	if [ ${INSTANCE_NAME} = "CHANGE_ME" ]; then
		echo "-----------------------------------------------------------"
		echo "Please update the file go.conf with your settings first !!!"
		echo "-----------------------------------------------------------"
		exit 41
	fi
	
	grep WELL_KNOWN_HTTP Dockerfile > /dev/null
	if [ ${?} = 0 ]; then
		echo "-----------------------------------------------------------"
                echo "Please run ./go update-ports prior to running build !!!"
                echo "-----------------------------------------------------------"
                exit 42
	fi

	create_required_directories
	get_required_files
	change_ports

	echo "User: ${IMAGE_USER}"
	echo "User Home: ${IMAGE_USER_HOME}"
	echo "User Shell: ${IMAGE_USER_SHELL}"

	docker build  --rm \
	--tag ${INSTANCE_NAME} \
	--build-arg UID=`id -u` \
	--build-arg GID=`id -g` \
	--build-arg USR=${IMAGE_USER} \
	--build-arg IMG_USR_HOME=${IMAGE_USER_HOME} \
	--build-arg IMG_USR_SHELL=${IMAGE_USER_SHELL} \
	--build-arg INSTALL_FROM=${INSTALL_FROM} \
	--build-arg COMPANY=${COMPANY} \
	--build-arg OS_UPDATES=${OS_UPDATES} \
	--build-arg EXTRA_PACKAGES=${EXTRA_PACKAGES} \
	--build-arg INSTANCE_NAME=${INSTANCE_NAME} \
	--build-arg USE_PROXY=${USE_PROXY} \
	--build-arg PROXY_URL=${PROXY_URL} \
	.


	copy_tomcat_apps
  ;;
  copy-files)
   	create_required_directories
   	get_required_files
  ;;
  update-ports)
	change_ports
  ;;
  *)
	usage
  ;;
esac

