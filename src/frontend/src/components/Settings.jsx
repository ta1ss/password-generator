import { useState,useEffect } from 'react';
import { getSettingFromUrl, setSettingInUrl} from '../utils/Urlparse.jsx';
import { isInputValid } from '../utils/Validation.jsx';
import PropTypes from 'prop-types';

Settings.propTypes = {
    settings: PropTypes.object.isRequired,
    setSettings: PropTypes.func.isRequired
};

function Settings({ settings, setSettings }) {
    const [minPasswordLength, setMinPasswordLength] = useState(getSettingFromUrl('minPasswordLength',15));
    const [maxPasswordLength, setMaxPasswordLength] = useState(getSettingFromUrl('maxPasswordLength',1000));
    const [minPasswordLengthLimit,setMinPasswordLengthLimit] = useState();
    const [maxPasswordLengthLimit,setMaxPasswordLengthLimit] = useState();

    // Set the limit of password length on initial load
    useEffect(() => {
        fetch('/api/v1/config/')
            .then(response => response.json())
            .then(data => setMinPasswordLengthLimit(data.MIN_PASSWORD_LENGTH))
            .catch(error => {
                console.error("Error fetching data:", error);
                setMinPasswordLengthLimit(15); // Set to default value
            });
            setMinPasswordLength(getSettingFromUrl('minPasswordLength', minPasswordLengthLimit));
        } ,[minPasswordLengthLimit]);

    useEffect(() => {
            fetch('/api/v1/config/')
            .then(response => response.json())
            .then(data => setMaxPasswordLengthLimit(data.MAX_PASSWORD_LENGTH))
            .catch(error => {
                console.error("Error fetching data:", error);
                setMaxPasswordLengthLimit(1000); // Set to default value
            });
            setMaxPasswordLength(getSettingFromUrl('maxPasswordLength', maxPasswordLengthLimit));
    },[maxPasswordLengthLimit]);

    const handleMinPasswordLengthInputChange = (event) => {
        const inputValue = event.target.value;
        setSettingInUrl('minPasswordLength', inputValue);
        setMinPasswordLength(inputValue);
        setSettings(prevSettings => ({ ...prevSettings, [minPasswordLength]: inputValue }));
    };

    const handleMaxPasswordLengthInputChange = (event) => {
        const inputValue = event.target.value;
        if (inputValue < minPasswordLength) {
            console.error("Max password length must be greater than min password length");
            return;
        }
        setSettingInUrl('maxPasswordLength', inputValue);
        setMaxPasswordLength(inputValue);
        setSettings(prevSettings => ({ ...prevSettings, [maxPasswordLength]: inputValue }));
    };

    useEffect(() => {
        const handlePopstate = () => {
                setMinPasswordLength(getSettingFromUrl('minPasswordLength', minPasswordLengthLimit));
                setMaxPasswordLength(getSettingFromUrl('maxPasswordLength', maxPasswordLengthLimit));
        };
        window.addEventListener('popstate', handlePopstate);
        return () => {
            window.removeEventListener('popstate', handlePopstate);
        };
    },[]);

    return (
        <div> {/* Wrap the two <p> elements inside a <div> */}
            <p>Minimum length of passwords: 
            <input
                type="number"
                id="minPasswordLength"
                name="minPasswordLength"
                min={minPasswordLengthLimit}
                max={maxPasswordLengthLimit}
                defaultValue={minPasswordLength}
                onChange={handleMinPasswordLengthInputChange}
                className={`form-control ${isInputValid (minPasswordLength,minPasswordLengthLimit,maxPasswordLengthLimit) ? "" : "is-invalid"}`}
                placeholder="Minimal length of passwords (15 - 1000)"
            />
            </p>
            <p>Maximum length of passwords: 
            <input
                type="number"
                id="maxPasswordLength"
                name="maxPasswordLength"
                min={minPasswordLength}
                max={maxPasswordLengthLimit}
                defaultValue={maxPasswordLengthLimit}
                onChange={handleMaxPasswordLengthInputChange}
                className={`form-control ${isInputValid (maxPasswordLength,minPasswordLength,maxPasswordLengthLimit) ? "" : "is-invalid"}`}
                placeholder="Maximal length of passwords (15 - 1000)"
            />
            </p>
        </div>
    );
}

export default Settings;
