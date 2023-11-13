import '../styles/App.css';
import PasswordTable from './PasswordTable.jsx';
import { useState, useEffect } from 'react';
import { NavLink, useLocation } from 'react-router-dom';

function App() {
    const [passwords, setPasswords] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [num, setNum] = useState("1");
    const [isInputValid, setIsInputValid] = useState(true);

    const [jsonLink, setJsonLink] = useState(`${window.location.origin}/json`);

    
    const location = useLocation();

    useEffect(() => {
        const handlePopstate = () => {
            const updatedNum = getNumFromURL();
            if (updatedNum) {
                setNum(updatedNum);
            }
        };

        window.addEventListener('popstate', handlePopstate);

        return () => {
            window.removeEventListener('popstate', handlePopstate);
        };
    }, []);

    useEffect(() => {
        const getNumFromURL = () => {
            const urlParams = new URLSearchParams(location.search);
            const numParam = urlParams.get('num');
            return numParam;
        };

        const initialNum = getNumFromURL();
        
        if (initialNum) {
            setNum(initialNum);
        }
    }, [location]);

    useEffect(() => {
        if (num && num.match(/^[0-9]*$/) && num >= 1 && num <= 1000) {
            setIsLoading(true);
            setIsInputValid(true);

            const newURL = `${window.location.origin}?num=${num}`;
            window.history.replaceState({}, '', newURL);

            fetch(`/json?num=${num}`)
                .then((response) => response.json())
                .then((data) => {
                    setPasswords(data);
                    setIsLoading(false);
                })
                .catch((error) => {
                    console.error("Error fetching data:", error);
                    setIsLoading(false);
                });
        } else {
            setPasswords([]);
            setIsInputValid(false);
        }
    }, [num]);

    const handleNumInputChange = (event) => {
        const inputValue = event.target.value;
        setNum(inputValue);

        if (inputValue.match(/^[0-9]*$/) && inputValue >= 1 && inputValue <= 1000) {
            setIsLoading(true);
            setIsInputValid(true);

            setJsonLink(`${window.location.origin}/json?num=${inputValue}`);
            fetch(`/json?num=${inputValue}`)
                .then((response) => response.json())
                .then((data) => {
                    setPasswords(data);
                    setIsLoading(false);
                })
                .catch((error) => {
                    console.error("Error fetching data:", error);
                    setIsLoading(false);
                });
        } else {
            setPasswords([]);
            setIsInputValid(false);
        }
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


            <div className="mt-3 ">
                {isLoading ? <p>Loading...</p> : <PasswordTable passwords={passwords} />}
            </div>
        </div>
    );
}


export default App;
