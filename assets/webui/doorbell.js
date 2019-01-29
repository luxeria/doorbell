window.onload = () => {
    function performRecaptchaCheck(siteKey) {
        const gr = "grecaptcha";
        return new Promise(resolve => {
            if (gr in window) {
                // recaptcha is already loaded
                resolve();
            } else {
                // inject recaptcha script tag dynamically
                const script = document.createElement("script");
                script.src = `https://www.google.com/recaptcha/api.js?render=${siteKey}`;
                script.onload = () => window[gr].ready(resolve);
                document.body.appendChild(script);
            }
        }).then(() => window[gr].execute(siteKey));
    }

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

        if (!resp.ok && maxTries > 1) {
            authToken.invalidate();
            return await ringDoorbell(authToken, maxTries - 1);
        }

        return resp.ok;
    }

    class AuthToken {
        constructor(recaptchaSiteKey, storageKey = "doorbell_jwt") {
            this.siteKey = recaptchaSiteKey;
            this.storagekey = storageKey;
        }

        async obtain() {
            let token = sessionStorage.getItem(this.storagekey);
            if (!token) {
                token = await performRecaptchaCheck(this.siteKey)
                    .then(authVerifyRecaptcha)
                    .then(r => r.token);
                sessionStorage.setItem(this.storagekey, token);
            }
            return token;
        }

        invalidate() {
            sessionStorage.removeItem(this.storagekey);
        }
    }

    const token = new AuthToken(REPATCHA_SITE_KEY);
    document
        .querySelector("button")
        .addEventListener("click", async (e) => {
            await ringDoorbell(token);
        });
};