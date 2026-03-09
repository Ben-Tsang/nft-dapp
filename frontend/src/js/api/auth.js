import { post, get } from "./request";

const authApi = {
  login(data) {
    return post("/auth/login", data);
  },

  nonce() {
    return get("/auth/nonce", null)
      .then((resData) => {
        console.log("resData:", resData);
        return resData.nonce || (resData.data && resData.data.nonce);
      })
      .catch((error) => {
        console.error("获取 nonce 失败:", error);
        throw error;
      });
  },

  logout() {
    localStorage.removeItem("token");
    return post("/auth/logout");
  },

  getCaptcha() {
    return get("/auth/captcha", null, {
      skipAuth: true,
      responseType: "blob",
    });
  },
};

export default authApi;
