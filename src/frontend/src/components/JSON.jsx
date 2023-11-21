import React, { useState, useEffect } from 'react';
import { PasswordGeneratorClient } from './PassgenServiceClientPb';
import { GenerateRequest } from './passgen_pb'
import JSONPretty from "react-json-pretty";

function PasswordJsonDisplay() {
    const [num, setNum] = useState("1");
    const [isInputValid, setIsInputValid] = useState(true);
    const [passwords, setPasswords] = useState([]);
    const [isLoading, setIsLoading] = useState(true);

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
        } else {
            setIsInputValid(false);
        }
        }, [num]);
    useEffect(() => {
        if (isInputValid){
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
        }
    }, [num]);

    if (isLoading) {
        return <p>Loading passwords...</p>;
    }

    return (
        <JSONPretty id="json-pretty" data={passwords}>
        </JSONPretty>
    );
}
export default PasswordJsonDisplay;