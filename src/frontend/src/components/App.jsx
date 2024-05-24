import '../styles/App.css';
import { getSettingFromUrl, setSettingInUrl } from '../utils/Urlparse.jsx';
import PasswordTable from './PasswordTable.jsx';
import Settings from './Settings.jsx';
import { useState, useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import { Collapse, Button } from 'reactstrap';

function App() {
    const [passwords, setPasswords] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [isError, setIsError] = useState(false);
    const [num, setNum] = useState(getSettingFromUrl('num', 1));
    const [isInputValid, setIsInputValid] = useState(true);
    const [settings, setSettings] = useState({});

    const [jsonLink, setJsonLink] = useState(`${window.location.origin}/json`);


    const [isOpen, setIsOpen] = useState(false);

    const toggle = () => setIsOpen(!isOpen);

    useEffect(() => {
        const handlePopstate = () => {
            // setNum(getSettingFromUrl('num', 1));
        };

        window.addEventListener('popstate', handlePopstate);

        return () => {
            window.removeEventListener('popstate', handlePopstate);
        };
    }, []);

    const updatePasswords = (inputValue) => {
        setIsError(false);

        if (String(inputValue).match(/^[0-9]*$/) && inputValue >= 1 && inputValue <= 1000) {
            setIsLoading(true);
            setIsInputValid(true);

            // Create a URLSearchParams object from the current inputs
            const urlParams = new URLSearchParams(window.location.search);
            // Create the fetch URL
            const fetchUrl = `/api/v1/passwords?${urlParams.toString()}`;

            // TODO: Check for value errors
            
            setJsonLink(`${window.location.origin}${fetchUrl}`);
            fetch(fetchUrl)
                .then((response) => {
                    if (!response.ok) {
                        throw new Error(`Error fetching password list! Status: ${response.status}`);
                    }
                    return response.json();
                })
                .then((data) => {
                    setPasswords(data);
                    setIsLoading(false);
                })
                .catch((error) => {
                    console.error("Error fetching data:", error);
                    setIsLoading(false);
                    setIsError(true);
                });
        } else {
            setPasswords([]);
            setIsInputValid(false);
            setIsLoading(true);
        }
    };

    useEffect(() => {
        // This code will run whenever the settings change

        setSettingInUrl('num', num);
        updatePasswords(num);
    }, [settings,num]); // Pass the settings as a dependency to useEffect

    const handleNumInputChange = (event) => {
        const inputValue = event.target.value;
        setNum(inputValue);
    };

    const handleFormSubmit = (event) => {
        event.preventDefault();
    };

    return (
        <div className="container mt-5 ">
            <div className="header text-center">
                <h1>Password Generator</h1>
                <div className="nav-links">
                    <NavLink to="/help" className="custom-link">
                        Help
                    </NavLink>
                    <span className="vertical-line"></span>
                    <a id="jsonLink" href={jsonLink} className="custom-link">
                        JSON
                    </a>
                    <span className="vertical-line"></span>
                    <a id="apiLink" href="/swagger/index.html" className="custom-link">
                        API
                    </a>
                </div>
            </div>

            <hr className="my-0" />

            <form id="numForm" className="password-input" onSubmit={handleFormSubmit}>
                <div className="row">
                    <div className="col-md-4">
                        <div className="form-group">
                            <input
                                type="number"
                                id="num"
                                name="num"
                                min="1"
                                max="1000"
                                value={num}
                                onChange={handleNumInputChange}
                                className={`form-control ${isInputValid ? "" : "is-invalid"}`}
                                placeholder="Number of Passwords (1 - 1000)"
                            />
                        </div>
                    </div>
                </div>
            </form>
            <div className="mt-3">
                <Button color="primary" onClick={toggle} style={{ marginBottom: '1rem' }}>Toggle Settings</Button>
                <Collapse isOpen={isOpen}>
                    <Settings settings={settings} setSettings={setSettings}/>
                </Collapse>
            </div>
            <div className="mt-3 ">
                {isError ? <p>Error occurred...</p> : isLoading ? <p>Loading...</p> : <PasswordTable passwords={passwords} />}
            </div>
        </div>
    );
}


export default App;
