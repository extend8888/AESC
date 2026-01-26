const { ethers } = require("hardhat");

async function main() {
  const [deployer] = await ethers.getSigners();
  
  console.log("Deploying EIP3009Token with account:", deployer.address);
  console.log("Account balance:", (await ethers.provider.getBalance(deployer.address)).toString());
  
  const tokenName = process.env.TOKEN_NAME || "Test USDT";
  const tokenSymbol = process.env.TOKEN_SYMBOL || "TUSDT";
  // Default: 1 billion tokens with 18 decimals = 1e27
  // This ensures sufficient balance for testing (relay_fee_per_tx = 0.01 token = 1e16)
  const initialSupply = process.env.INITIAL_SUPPLY || "1000000000000000000000000000";
  
  console.log("Token Name:", tokenName);
  console.log("Token Symbol:", tokenSymbol);
  console.log("Initial Supply:", initialSupply);
  
  // Get the contract factory
  const EIP3009Token = await ethers.getContractFactory("EIP3009Token");
  
  // Deploy the contract
  console.log("Deploying contract...");
  const token = await EIP3009Token.deploy(tokenName, tokenSymbol, initialSupply);
  
  // Wait for deployment
  await token.waitForDeployment();
  const tokenAddress = await token.getAddress();
  
  console.log("EIP3009Token deployed to:", tokenAddress);
  
  // Verify deployment
  const name = await token.name();
  const symbol = await token.symbol();
  const decimals = await token.decimals();
  const totalSupply = await token.totalSupply();
  const domainSeparator = await token.DOMAIN_SEPARATOR();
  
  console.log("\n=== Deployed Token Info ===");
  console.log("Name:", name);
  console.log("Symbol:", symbol);
  console.log("Decimals:", decimals.toString());
  console.log("Total Supply:", totalSupply.toString());
  console.log("DOMAIN_SEPARATOR:", domainSeparator);
  console.log("Deployer Balance:", (await token.balanceOf(deployer.address)).toString());
  
  // Write deployed address to file for test scripts
  const fs = require("fs");
  const deployInfo = {
    address: tokenAddress,
    name: name,
    symbol: symbol,
    decimals: Number(decimals),
    domainSeparator: domainSeparator,
    deployer: deployer.address,
    chainId: (await ethers.provider.getNetwork()).chainId.toString()
  };
  
  fs.writeFileSync("deployed_token.json", JSON.stringify(deployInfo, null, 2));
  console.log("\nDeployed info saved to deployed_token.json");
  
  return deployInfo;
}

main()
  .then((info) => {
    console.log("\n✅ Deployment successful!");
    process.exit(0);
  })
  .catch((error) => {
    console.error("❌ Deployment failed:", error);
    process.exit(1);
  });

