import '../styles/Help.css';
import { NavLink } from 'react-router-dom';

function Help() {
    const currentHost = window.location.origin;

    return (
        <div className="container mt-5">
            <div className="header text-center">
                <h1>Help</h1>
                <div className="nav-links">
                    <NavLink to="/" className="custom-link">
                        Home
                    </NavLink>
                </div>
            </div>

            <hr className="my-0" />

            <ul className=" password-input list-group">
                <div className="list-group-item" >
                    <p>Generate multiple passwords (1 - 1000) by adding <code>{`?num=`}</code> to your URL:</p>
                    <span className="link-icon ms-3">Example: 🔗</span> <a className="custom-link-url" href={`${currentHost}?num=50`}><code>{`${currentHost}?num=50`}</code></a>
                </div>

                <div className="list-group-item" >
                    <p>Generate plain JSON output by adding <code>{`/json`}</code> to your URL:</p>
                    <span className="link-icon ms-3">Example: 🔗</span> <a className="custom-link-url" href={`${currentHost}/json`}><code>{`${currentHost}/json`}</code></a>
                </div>

                <div className="list-group-item" >
                    <p>Generate multiple JSON passwords by adding <code>{`/json?num=`}</code> to your URL:</p>
                    <span className="link-icon ms-3">Example: 🔗</span> <a className="custom-link-url" href={`${currentHost}/json?num=50`}><code>{`${currentHost}/json?num=50`}</code></a>
                </div>
            </ul>
        </div>

    );
}

export default Help;