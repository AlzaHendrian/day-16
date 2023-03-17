function getData(e) {
    e.preventDefault();
    let name = document.getElementById('name').value
    let email = document.getElementById('email').value
    let telp = document.getElementById('telp').value
    let subject = document.getElementById('subject').value
    let message = document.getElementById('message').value

    if (name == "") {
        alert("please, input your name!")
    } else if (email == "") {
        alert("please, input your email!")
    } else if (telp == "") {
        alert("please, input your number phone!")
    } else if (subject == "") {
        alert("please, input your subject!")
    } else if (message == "") {
        alert("please, input your message!")
    } else {

        const defaultEmail = "alzahendriaan@gmail.com"

        let mailTo = document.createElement('a')
        mailTo.href = `mailto:${defaultEmail}?subject=${subject}&body=Hi my name is ${name}, I want to ${message} Please call me back in my number : ${telp} , Thankyou.`
        mailTo.click()

        let audience = {
            name,
            email,
            telp,
            subject,
            message,
        }

        console.log(audience)

    }

}