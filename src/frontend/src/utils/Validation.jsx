
function isInputValid(inputValue, min, max) {
    inputValue = String(inputValue);
    return inputValue.match(/^[0-9]*$/) && inputValue >= min && inputValue <= max;
}

export { isInputValid };