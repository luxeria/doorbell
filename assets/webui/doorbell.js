import * as config from "./config.js";

async function authVerifyRecaptcha(captchaResponse) {
    const resp = await fetch("/auth/recaptcha", {
        method: "POST",
        headers: {
            "Accept": "application/json",
            "Content-Type": "application/json",
        },
        body: JSON.stringify({response: captchaResponse})
    });

    if (!resp.ok) {
        throw new Error(resp.statusText);
    }

    return await resp.json()
}

async function ringDoorbell(authToken, maxTries = 2) {
    const resp = await fetch("/ring", {
        method: "POST",
        headers: {
            "Accept": "application/json",
            "Authorization": `Bearer ${await authToken.obtain()}`,
        }
    });

    if (!resp.ok) {
        if (maxTries > 1) {
            authToken.invalidate();
            await ringDoorbell(authToken, maxTries - 1);
        } else {
            const message = await resp.json()
                .then(msg => msg.error)
                .catch(() => resp.statusText);
            throw new Error(message)
        }
    }

    return resp.ok;
}

class AuthToken {
    constructor(conf) {
        this.recaptchaSiteKey = conf.recaptchaSiteKey;
        this.jwtStoragekey = conf.jwtStorageKey;
    }

    async obtain() {
        let token = sessionStorage.getItem(this.jwtStoragekey);
        if (!token) {
            token = await grecaptcha.execute(this.recaptchaSiteKey)
                .then(authVerifyRecaptcha)
                .then(r => r.token);
            sessionStorage.setItem(this.jwtStoragekey, token);
        }
        return token;
    }

    invalidate() {
        sessionStorage.removeItem(this.jwtStoragekey);
    }
}

const userToken = new AuthToken(config);

window.addEventListener("load", () => {
    document
        .querySelector(".doorbell")
        .addEventListener("click", async () => {
            await ringDoorbell(userToken);
        });
});