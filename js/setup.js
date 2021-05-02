const dhcpSlider = document.getElementById("dhcp-slider")
const leftButton = document.getElementById("left-button")
const middleButton = document.getElementById("middle-button")
const rightButton = document.getElementById("right-button")
const passwordField = document.getElementById("password")
const ipaddress = document.getElementById("ipaddress")
const gateway = document.getElementById("gateway")
const server = document.getElementById("server")
const mask = document.getElementById("mask")

const Keyboard = {
    elements: {
        main: null,
        keysContainer: null,
        keys: []
    },
    eventHandlers: {
        oninput: null,
        onclose: null
    },
    properties: {
        value: "",
        capsLock: false
    },
    init() {
        this.elements.main = document.createElement("div");
        this.elements.keysContainer = document.createElement("div");
        this.elements.main.classList.add("keyboard", "keyboard--hidden");
        this.elements.keysContainer.classList.add("keyboard__keys");
        this.elements.keysContainer.appendChild(this._createKeys());
        this.elements.keys = this.elements.keysContainer.querySelectorAll(".keyboard__key");
        this.elements.main.appendChild(this.elements.keysContainer);
        document.body.appendChild(this.elements.main);
        document.querySelectorAll(".use-keyboard-input").forEach(element => {
            element.addEventListener("focus", () => {
                this.open(element.value, currentValue => {
                    element.value = currentValue;
                });
            });
        });
    },
    _createKeys() {
        const fragment = document.createDocumentFragment();
        const keyLayout = [
            "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "backspace",
            "q", "w", "e", "r", "t", "y", "u", "i", "o", "p",
            "a", "s", "d", "f", "g", "h", "j", "k", "l", "enter",
            "z", "x", "c", "v", "b", "n", "m", ".", ":", "/",
        ];
        keyLayout.forEach(key => {
            const keyElement = document.createElement("button");
            const insertLineBreak = ["backspace", "p", "enter", "/"].indexOf(key) !== -1;
            keyElement.setAttribute("type", "button");
            keyElement.classList.add("keyboard__key");
            switch (key) {
                case "backspace":
                    keyElement.classList.add("keyboard__key--wide");
                    keyElement.innerHTML = "⌫";
                    keyElement.addEventListener('touchstart', () => {
                        this.properties.value = this.properties.value.substring(0, this.properties.value.length - 1);
                        this._triggerEvent("oninput");
                    });
                    break;
                case "enter":
                    keyElement.classList.add("keyboard__key--wide");
                    keyElement.innerHTML = "↵";
                    keyElement.addEventListener('touchstart', () => {
                        if (sessionStorage.getItem("selection") === "password") {
                            let password = this.properties.value
                            let data = {
                                password: password
                            };
                            fetch("/password", {
                                method: "POST",
                                body: JSON.stringify(data)
                            }).then((response) => {
                                response.text().then(function (data) {
                                    let result = JSON.parse(data);
                                    if (result["Result"] === "ok") {
                                        document.getElementById("password").hidden = true
                                        if (!dhcpSlider.checked) {
                                            document.getElementById("ipaddress").disabled = false
                                            document.getElementById("gateway").disabled = false
                                            document.getElementById("mask").disabled = false

                                        } else {
                                            Keyboard.close()
                                        }
                                        middleButton.disabled = false
                                        middleButton.style.pointerEvents = "auto"
                                        rightButton.disabled = false
                                        rightButton.style.pointerEvents = "auto"
                                        document.getElementById("server").disabled = false
                                        document.getElementById("dhcp-slider").disabled = false
                                    }
                                })
                            }).catch(() => {
                            });
                        }
                        this._triggerEvent("oninput");
                    });
                    break;
                default:
                    keyElement.textContent = key.toLowerCase();
                    keyElement.addEventListener('touchstart', () => {
                        this.properties.value += this.properties.capsLock ? key.toUpperCase() : key.toLowerCase();
                        this._triggerEvent("oninput");
                    });
                    break;
            }
            fragment.appendChild(keyElement);
            if (insertLineBreak) {
                fragment.appendChild(document.createElement("br"));
            }
        });
        return fragment;
    },
    _triggerEvent(handlerName) {
        if (typeof this.eventHandlers[handlerName] == "function") {
            this.eventHandlers[handlerName](this.properties.value);
        }
        let elem = document.getElementById('server');
        elem.scrollLeft = elem.scrollWidth;
    },
    open(initialValue, oninput, onclose) {
        this.properties.value = initialValue || "";
        this.eventHandlers.oninput = oninput;
        this.eventHandlers.onclose = onclose;
        this.elements.main.classList.remove("keyboard--hidden");
    },
    close() {
        this.properties.value = "";
        this.eventHandlers.oninput = oninput;
        this.eventHandlers.onclose = onclose;
        this.elements.main.classList.add("keyboard--hidden");
    }
};

window.addEventListener("DOMContentLoaded", function () {
    Keyboard.init();
});

dhcpSlider.addEventListener('change', function () {
    if (dhcpSlider.checked) {
        document.getElementById("ipaddress").disabled = true
        document.getElementById("gateway").disabled = true
        document.getElementById("mask").disabled = true
    } else {
        document.getElementById("ipaddress").disabled = false
        document.getElementById("gateway").disabled = false
        document.getElementById("mask").disabled = false
    }
}, false);

leftButton.addEventListener('touchstart', function () {
    leftButton.style.border = "2px solid red"
    middleButton.style.border = "2px solid white"
    rightButton.style.border = "2px solid white"
    window.open("/", "_self")
}, false);

leftButton.addEventListener('touchstart', function () {
    leftButton.style.border = "2px solid red"
    middleButton.style.border = "2px solid white"
    rightButton.style.border = "2px solid white"
    window.open("/", "_self")
}, false);


middleButton.addEventListener('touchend', function () {
    middleButton.style.border = "2px solid white"
    middleButton.blur()
    setTimeout(() => {
        middleButton.blur()
    }, 10);
})

rightButton.addEventListener('touchend', function () {
    rightButton.style.border = "2px solid white"
    rightButton.blur()
    setTimeout(() => {
        rightButton.blur()
    }, 10);
})

middleButton.addEventListener('touchstart', function () {
    middleButton.blur()
    if (!dhcpSlider.checked) {
        leftButton.style.border = "2px solid white"
        middleButton.style.border = "2px solid red"
        rightButton.style.border = "2px solid white"
        ipaddress.value = ""
        mask.value = ""
        gateway.value = ""
        server.value = ""
    } else {
        leftButton.style.border = "2px solid white"
        middleButton.style.border = "2px solid red"
        rightButton.style.border = "2px solid white"
        server.value = ""
    }
    middleButton.blur()
}, false);

function checkInputData() {
    let result = false
    let ipResult = false
    let maskResult = false
    let gatewayResult = false
    if (/^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(ipaddress.value)) {
        ipResult = true
        ipaddress.style.border = "1px solid white"
    } else {
        ipaddress.style.border = "1px solid red"
    }
    if (/^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(mask.value)) {
        maskResult = true
        mask.style.border = "1px solid white"
    } else {
        mask.style.border = "1px solid red"
    }
    if (/^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(gateway.value)) {
        gatewayResult = true
        mask.style.border = "1px solid white"
    } else {
        gateway.style.border = "1px solid red"
    }
    if (ipResult && maskResult && gatewayResult) {
        ipaddress.style.border = "1px solid white"
        gateway.style.border = "1px solid white"
        mask.style.border = "1px solid white"
        result = true
    }
    return result;
}

rightButton.addEventListener('touchstart', function () {
    leftButton.style.border = "2px solid white"
    middleButton.style.border = "2px solid white"
    rightButton.style.border = "2px solid red"
    rightButton.blur()
    if (dhcpSlider.checked) {
        document.getElementById("ipaddress").disabled = true
        document.getElementById("gateway").disabled = true
        document.getElementById("mask").disabled = true
        let data = {
            password: "3600",
            server: server.value,
        };
        fetch("/dhcp", {
            method: "POST",
            body: JSON.stringify(data)
        }).then(() => {
            window.open("/", "_self")
        }).catch(() => {
        });
    } else {
        let resultOk = checkInputData();
        if (resultOk) {
            let data = {
                password: "3600",
                ipaddress: ipaddress.value,
                mask: mask.value,
                gateway: gateway.value,
                server: server.value,
            };
            fetch("/static", {
                method: "POST",
                body: JSON.stringify(data)
            }).then(() => {
                window.open("/", "_self")
            }).catch(() => {
            });
        }
    }
}, false);

passwordField.addEventListener('touchstart', function () {
    sessionStorage.setItem("selection", "password")
}, false);

ipaddress.addEventListener('touchstart', function () {
    sessionStorage.setItem("selection", "ipaddress")
}, false);

server.addEventListener('touchstart', function () {
    sessionStorage.setItem("selection", "server")
}, false);

gateway.addEventListener('touchstart', function () {
    sessionStorage.setItem("selection", "gateway")
}, false);

mask.addEventListener('touchstart', function () {
    sessionStorage.setItem("selection", "mask")
}, false);

