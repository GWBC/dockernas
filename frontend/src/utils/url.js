import storage from "@/utils/storage"

export const getInstanceWebUrl = (instanceParam, portParamItem) => {
    if (instanceParam.networkMode === "host") {
        return "http://" + window.location.hostname + ":" + portParamItem.key;
    }
    if (portParamItem.value === "" || instanceParam.networkMode === "nobund") {
        return null;
    }
    // return window.location.protocol + "//"+window.location.hostname+":"+port
    return "http://" + window.location.hostname + ":" + portParamItem.value
}

export const getInstanceWebUrl2 = (instance, instanceParam, portParamItem) => {
    let hostname = instance.dockerSvrIP;
    if (hostname.lenght == 0)
    {
        hostname = window.location.hostname;
    }

    if (instanceParam.networkMode === "host") {
        return "http://" + hostname + ":" + portParamItem.key;
    }
    if (portParamItem.value === "" || instanceParam.networkMode === "nobund") {
        return null;
    }
    // return window.location.protocol + "//"+hostname+":"+port
    return "http://" + hostname+ ":" + portParamItem.value
}

export const getInstancePortText = (instanceParam, portParamItem) => {
    if (instanceParam.networkMode === "host") {
        return portParamItem.key;
    }

    if (portParamItem.value === "" || instanceParam.networkMode === "nobund") {
        return portParamItem.key + " -> ";
    }

    return portParamItem.key + " -> " + portParamItem.value;
}

export const splitRouterPathByIndex = (router, index) => {
    return router.split("/").slice(0, index).join("/")
}

export const getFirstHttpPortUrl = (instanceParam) => {
    if (instanceParam.networkMode === "nobund") {
        return null
    }
    for (let param of instanceParam.portParams) {
        if (param.protocol === 'http') {
            if (instanceParam.networkMode !== "host" && instanceParam.value === "") {
                continue
            }
            return getInstanceWebUrl(instanceParam, param)
        }
    }
    return null;
}

export const getIconUrl = (url) => {
    return url + "&token=" + storage.get("token");
}