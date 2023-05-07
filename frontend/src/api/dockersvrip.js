import http from '@/utils/request'

export const getDockerSvrInfos = () => {
    return http.get("/api/getdockersvrinfos")
};

export const addDockerSvrInfo = (info) => {
    return http.post("/api/adddockersvrinfo", info)
};

export const editDockerSvrInfo = (info) => {
    return http.post("/api/updatedockersvrinfo", info)
};

export const delDockerSvrInfo = (info) => {
    return http.post("/api/deletedockersvrinfo", info)
};



