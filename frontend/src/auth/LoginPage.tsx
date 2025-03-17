import React, { useContext, useState } from 'react';
import { Link, useNavigate, NavigateFunction, useSearchParams } from 'react-router-dom';
import { FormError } from '../FormError';
import { AuthContext, AuthContextType, LoginFormValidationErrors } from './AuthContext';

// When redirected to login can use ?referrer=/somePage to redirect after successful login attempt

async function handleFormSubmit(auth: AuthContextType, formData: FormData, navigate: NavigateFunction, setLoading: React.Dispatch<React.SetStateAction<boolean>>, setValidationErrors: React.Dispatch<React.SetStateAction<LoginFormValidationErrors>>) {
  setLoading(true)
  const redirectUrl = formData.get("referrer") as string || "/"
  const loginData = {
    username: formData.get("username") as string,
    password: formData.get("password") as string,
  }

  const loginCallback = (success: boolean, responseData: LoginFormValidationErrors | undefined) => {
    if (success) {
      navigate(redirectUrl)
    } else {
      if (responseData != undefined) {
        setValidationErrors(responseData)
      }
    }
  }

  auth.login(loginData, loginCallback)
  setLoading(false)
}

function LoginForm() {
  const [loading, setLoading] = useState(false)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [remember, setRemember] = useState(false)
  const [searchParams, _setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const [validationErrors, setValidationErrors] = useState({
    username: "",
    password: "",
  })
  const auth = useContext(AuthContext)
  
  return (
    <form method="post" action={(formData) => {if (!loading) {handleFormSubmit(auth, formData, navigate, setLoading, setValidationErrors)}}}>
      <div className='formGroup'>
        <label htmlFor="username">Username</label>
        <input name="username" type="text" required={true} value={username} onChange={(event) => setUsername(event.target.value)}/>
        <FormError errorMessage={validationErrors.username} />
      </div>
      <div className='formGroup'>
        <label htmlFor="password">Password</label>
        <input name="password" type="password" required={true} value={password} onChange={(event) => setPassword(event.target.value)}/>
      </div>
      <button className='signInButton'>SIGN IN</button>
      <label>
        <input type="checkbox" style={{marginLeft: "0"}} checked={remember} onChange={() => setRemember(!remember)}/>
        Keep me logged in
      </label>
      <input className="hidden" name="referrer" type="text" required={false} value={searchParams.get("referrer") || ""}/>
    </form>
  )
}

function LoginOptions() {
  const [searchParams, _setSearchParams] = useSearchParams()
  const referrer = searchParams.get("referrer")
  let registerUrl = "/register"
  if (referrer != null) {
    registerUrl += `?referrer=${referrer}`
  }
  return (
    <div>
      <div className="loginOptions">
        <Link to={registerUrl}>Register</Link>
        <Link to="/resetPassword" style={{marginLeft: "auto"}}>Password Reset</Link>
      </div>
    </div>
  )
}

export function LoginPage() {
  return (
    <div className="loginTile">
      <h1>Sign In</h1>
      <LoginForm />
      <LoginOptions />
    </div>
  )
}