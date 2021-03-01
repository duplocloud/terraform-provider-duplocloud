import requests
import pprint
from os import environ
import time
import sys
from logger import logging

logger = None

duplo_engine =  environ.get('DUPLO_EP') + '/subscriptions/' + environ.get('TENANT_ID')

def deploy_new_service(g_serviceName, aInImage, aInHeaders):
    data = {
                "Name": g_serviceName,
                "Image": aInImage 
            }
    logger.info(data)
    postUrl = duplo_engine + '/ReplicationControllerChange'
    response = requests.post(postUrl, json=data, headers=aInHeaders)
    response.raise_for_status()
    print('Updated the service')
	
def check_containers_running(g_serviceName, aInHeaders):	
    getUrl = duplo_engine + '/GetPods'
    response = requests.get(getUrl, headers=aInHeaders)
    allOk = False
    for pod in response.json():
        if pod["Name"].lower() != g_serviceName.lower():
            continue
        if pod["CurrentStatus"] != 1:
            logger.info('Service %s at least one container is not running, current status %s', g_serviceName, pod["CurrentStatus"])
            return False
    
        logger.info("All containers in service %s are running", g_serviceName)
        return True

def validate_service(g_serviceName, aInHeaders):
    allOk = False
    for x in range(240):
        time.sleep(5)
        allOk = check_containers_running(g_serviceName, aInHeaders)
        if allOk:
            break

    if not allOk:
        err = 'Deploy failed, some containers failed to get into running state '
        raise Exception(err) 	

	
def setup_custom_logger(name):
    formatter = logging.Formatter(fmt='%(asctime)s %(levelname)-8s %(message)s',
                                  datefmt='%Y-%m-%d %H:%M:%S')
    logger = logging.getLogger(name)
    logger.setLevel(logging.DEBUG)
    return logger

if __name__ == '__main__':
    logger = setup_custom_logger('CustomDeploy')
    jheaders = {"Content-type": "application/json", "Authorization": "Bearer " + environ.get('DUPLO_SSO_TOKEN')}
    deploy_new_service(sys.argv[1], sys.argv[2], jheaders)
    validate_service(sys.argv[1], jheaders)
