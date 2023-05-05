import storage from "@/utils/storage"

export const getInstanceWebUrl = (instance, instanceParam, portParamItem) => {
    let hostname = window.location.hostname;

    if (instance){
        if (instance.dockerSvrIP && instance.dockerSvrIP.lenght != 0) {
            hostname =instance.dockerSvrIP;
        }
    }
   
    if (instanceParam.networkMode === "host") {
        return "http://" + hostname + ":" + portParamItem.key;
    }

    if (portParamItem.value === "" || instanceParam.networkMode === "nobund") {
        return null;
    }

    return "http://" + hostname+ ":" + portParamItem.value
}

export const getInstancePortText = (instanceParam, portParamItem) => {
    if (instanceParam.networkMode === "host") {
        return portParamItem.key;
    }

    if (portParamItem.value === "" || instanceParam.networkMode === "nobund") {
        return " -> " + portParamItem.key;
    }

    return portParamItem.value + " -> " + portParamItem.key;
}

export const splitRouterPathByIndex = (router, index) => {
    return router.split("/").slice(0, index).join("/")
}

export const getFirstHttpPortUrl = (instance, instanceParam) => {
    if (instanceParam.networkMode === "nobund") {
        return null
    }
    for (let param of instanceParam.portParams) {
        if (param.protocol === 'http') {
            if (instanceParam.networkMode !== "host" && instanceParam.value === "") {
                continue
            }
            return getInstanceWebUrl(instance, instanceParam, param)
        }
    }
    return null;
}

export const getIconUrl = (url) => {
    return url + "&token=" + storage.get("token");
}