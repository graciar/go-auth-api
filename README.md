# 🛡️ Go Authentication API (Gin + MongoDB + JWT)

A simple authentication API built with **Go (Gin framework)**, **MongoDB**, and **JWT**.  
This project demonstrates secure user authentication, including registration, login, and password reset using **SendGrid OTP verification**.

---

## 🚀 Features

- 🔐 **Register** 
- 🔑 **Login**   
- 🔄 **Forgot Password** 
- 📨 **Verify OTP**  
- 🔁 **Reset Password** 
- 🧠 **JWT-based Authentication**

---

## 🧰 Tech Stack

- **Backend:** Go + Gin  
- **Database:** MongoDB  
- **Authentication:** JWT  
- **Email Service:** SendGrid  

---

## ⚙️ Setup 

### 1️⃣ Clone the Repository

```
git clone https://github.com/graciar/go-auth-api.git
cd go-auth-api
```

### 2️⃣ Create .env File
Copy .env.example to .env and fill in your environment values:
```
cp .env.example .env
```

3️⃣ Install Dependencies
```
go mod tidy
```

4️⃣ Run the Server
```
go run main.go
```
