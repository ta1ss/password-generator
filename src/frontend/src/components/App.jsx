import '../styles/App.css';
import PasswordTable from './PasswordTable.jsx';
import React, { useState, useEffect } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { PasswordGeneratorClient } from './PassgenServiceClientPb';
import { GenerateRequest } from './passgen_pb'

function App() {
    const [passwords, setPasswords] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [num, setNum] = useState("1");
    const [isInputValid, setIsInputValid] = useState(true);    
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

            const client = new PasswordGeneratorClient('');
            let request = new GenerateRequest();
            request.setLen(parseInt(num));


            client.getPasswords(request, {}, (err, response) => {
                if (err) {
                    console.error("Error calling GetPasswords:", err.message);
                    setIsLoading(false);
                } else {
                    const passwords = response.getPasswordsList().map(p => ({
                        xkcd: p.getXkcd(),
                        original: p.getOriginal(),
                        length: p.getLength()
                    }));
                    setPasswords(passwords);
                    setIsLoading(false);
                }
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

            const client = new PasswordGeneratorClient('');
            let request = new GenerateRequest();
            request.setLen(parseInt(num));


            client.getPasswords(request, {}, (err, response) => {
                if (err) {
                    console.error("Error calling GetPasswords:", err.message);
                    setIsLoading(false);
                } else {
                    const passwords = response.getPasswordsList().map(p => ({
                        xkcd: p.getXkcd(),
                        original: p.getOriginal(),
                        length: p.getLength()
                    }));
                    setPasswords(passwords);
                    setIsLoading(false);
                }
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
                    <NavLink to="/json" className="custom-link">
                        JSON
                    </NavLink>
                    <span className="vertical-line"></span>
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
