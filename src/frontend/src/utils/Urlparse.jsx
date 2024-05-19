function getSettingFromUrl(settingName,defaultValue) {
    const urlParams = new URLSearchParams(window.location.search);
    const settingValue = urlParams.get(settingName);
    return settingValue || defaultValue;
}

function setSettingInUrl(settingName, settingValue) {
    const url = new URL(window.location);
    url.searchParams.set(settingName, settingValue);
    window.history.pushState({}, '', url);
}

export { getSettingFromUrl, setSettingInUrl};