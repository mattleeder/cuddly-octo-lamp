import React, { useContext, useState } from 'react';
import { useNavigate, NavigateFunction, useSearchParams } from 'react-router-dom';
import { FormError } from '../FormError';
import { AuthContext, AuthContextType, RegisterFormValidationErrors } from './AuthContext';

// When redirected to login can use ?referrer=/somePage to redirect after successful login attempt

async function handleFormSubmit(auth: AuthContextType, formData: FormData, navigate: NavigateFunction, setLoading: React.Dispatch<React.SetStateAction<boolean>>, setValidationErrors: React.Dispatch<React.SetStateAction<RegisterFormValidationErrors>>) {
  setLoading(true)
  const redirectUrl = formData.get("referrer") as string || "/"
  const registerData = {
    username: formData.get("username") as string,
    password: formData.get("password") as string,
    email: formData.get("email") as string || undefined,
    rememberMe: formData.get("rememberMe") == "true" ? true : false,
  }

  console.log(registerData)

  const registerCallback = (success: boolean, responseData: RegisterFormValidationErrors | undefined) => {
    if (success) {
      navigate(redirectUrl)
    } else {
      if (responseData != undefined) {
        setValidationErrors(responseData)
      }
    }
  }

  auth.register(registerData, registerCallback)
  setLoading(false)
}

function RegisterForm() {
  const [loading, setLoading] = useState(false)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [email, setEmail] = useState("")
  const [remember, setRemember] = useState(false)
  const [searchParams, _setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const [validationErrors, setValidationErrors] = useState({
    username: "",
    password: "",
    email: ""
  })
  const auth = useContext(AuthContext)
  
  return (
    <form method="post" action={(formData) => {if (!loading) {handleFormSubmit(auth, formData, navigate, setLoading, setValidationErrors)}}}>
      <div className='formGroup'>
        <label htmlFor="username">Username</label>
        <input name="username" type="text" required={true} value={username} onChange={(event) => setUsername(event.target.value)} />
        <FormError errorMessage={validationErrors.username} />
      </div>
      <div className='formGroup'>
        <label htmlFor="password">Password</label>
        <input name="password" type="password" required={true} value={password} onChange={(event) => setPassword(event.target.value)}/>
      </div>
      <div className='formGroup'>
        <label htmlFor="email">Email (Optional - For password reset)</label>
        <input name="email" type="email" required={false} value={email} onChange={(event) => setEmail(event.target.value)}/>
      </div>
      <button className={`signInButton${loading ? " disabled" : ""}`}>REGISTER</button>
      <label>
        <input type="checkbox" style={{marginLeft: "0"}} checked={remember} onChange={() => setRemember(!remember)}/>
        Keep me logged in
      </label>
      <input className="hidden" name="referrer" type="text" required={false} value={searchParams.get("referrer") || ""}/>
    </form>
  )
}

export function RegisterPage() {
  return (
    <div className="registerTile">
      <h1>Register</h1>
      <RegisterForm />
    </div>
  )
}
