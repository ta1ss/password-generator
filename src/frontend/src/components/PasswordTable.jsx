import React, { useState } from 'react';


function PasswordTable({ passwords }) {
    const [highlightedColumn, setHighlightedColumn] = useState(null);
    const [copyNotification, setCopyNotification] = useState(null);

    const handleCopyClick = (text) => {
        const combinedText = text.join('\n');
        navigator.clipboard.writeText(combinedText)
            .then(() => {
                setCopyNotification('Text copied to clipboard!');
                setTimeout(() => {
                    setCopyNotification(null);
                }, 2000);
            })
            .catch((err) => {
                setCopyNotification(`Error coping text to clipboard: ${err}`)
            });
    };

    const generatedPasswords = passwords.map(password => password.Xkcd);
    const originalPasswords = passwords.map(password => password.Original);

    return (
        <div>
            {copyNotification && (
            <div className={`copy-notification ${copyNotification ? 'show' : ''}`}>
                {copyNotification}
            </div>
            )}
            <table>
                <thead>
                    <tr>
                        <th className="generated">
                            Password
                            <span
                                className="copy-icon"
                                onClick={() => handleCopyClick(generatedPasswords)}
                                onMouseEnter={() => setHighlightedColumn('generated')}
                                onMouseLeave={() => setHighlightedColumn(null)}
                            >
                                &#128203;
                            </span>
                        </th>
                        <th className="original">
                            Original
                            <span
                                className="copy-icon"
                                onClick={() => handleCopyClick(originalPasswords)}
                                onMouseEnter={() => setHighlightedColumn('original')}
                                onMouseLeave={() => setHighlightedColumn(null)}
                            >
                                &#128203;
                            </span>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {passwords.map((password, index) => (
                        <tr key={index}>
                            <td className={`generated ${highlightedColumn === 'generated' ? 'highlighted' : ''}`}>{password.xkcd}</td>
                            <td className={`original ${highlightedColumn === 'original' ? 'highlighted' : ''}`}>{password.original}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}


export default PasswordTable;