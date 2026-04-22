mkdir $Env:USERPROFILE/.aspnet/https

dotnet dev-certs https --clean
dotnet dev-certs https -ep $Env:USERPROFILE/.aspnet/https/Development.pfx -p c1bc6816-f70f-42e3-a71f-4ab75a294755
dotnet dev-certs https --trust
dotnet dev-certs https --check
