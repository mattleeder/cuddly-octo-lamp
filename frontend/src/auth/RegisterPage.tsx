import React, { useState } from 'react';
import { Link, useParams, useNavigate, NavigateFunction } from 'react-router-dom';
import { FormError } from '../FormError';

// When redirected to login can use ?referrer=/somePage to redirect after successful login attempt

interface RegisterFormValidationErrors {
  username: string
  password: string
  email: string
}

async function handleFormSubmit(formData: FormData, navigate: NavigateFunction, setLoading: React.Dispatch<React.SetStateAction<boolean>>, setValidationErrors: React.Dispatch<React.SetStateAction<RegisterFormValidationErrors>>) {
  setLoading(true)
  const url = import.meta.env.VITE_API_REGISTER_URL
  const redirectUrl = formData.get("referrer") as string || "/"

  try {
    const response = await fetch(url, {
      method: "POST",
      body: JSON.stringify({
        username: formData.get("username"),
        password: formData.get("password"),
        email: formData.get("email"),
      })
    })

    if (response.ok) {
      navigate(redirectUrl)
    } else {
      const validationErrors: RegisterFormValidationErrors = await response.json()
      setValidationErrors(validationErrors)
    }
  } catch (e) {
    console.error(e)
  } finally {
    setLoading(false)
  }
}

function RegisterForm() {
  const [loading, setLoading] = useState(false)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [email, setEmail] = useState("")
  const [remember, setRemember] = useState(false)
  const params = useParams()
  const navigate = useNavigate()
  const [validationErrors, setValidationErrors] = useState({
    username: "",
    password: "",
    email: ""
  })
  
  return (
    <form method="post" action={(formData) => {if (!loading) {handleFormSubmit(formData, navigate, setLoading, setValidationErrors)}}}>
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
      <input className="hidden" name="referrer" type="text" required={false} value={params.referrer}/>
    </form>
  )
}

function LoginOptions() {
  return (
    <div>
      <div className="loginOptions">
        <Link to="/register">Register</Link>
        <Link to="/resetPassword" style={{marginLeft: "auto"}}>Password Reset</Link>
      </div>
    </div>
  )
}

export function RegisterPage() {
  return (
    <div className="registerTile">
      <h1>Register</h1>
      <RegisterForm />
      <LoginOptions />
    </div>
  )
}