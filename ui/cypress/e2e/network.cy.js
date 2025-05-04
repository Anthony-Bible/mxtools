describe('Network Tools', () => {
  it('should allow submitting Ping', () => {
    cy.visit('/network');
    cy.get('#network-tool-select').select('ping');
    cy.get('input[type="text"]').type('8.8.8.8');
    cy.get('button[type="submit"]').click();
    cy.get('pre').should('exist');
  });

  it('should allow submitting Traceroute', () => {
    cy.visit('/network');
    cy.get('#network-tool-select').select('traceroute');
    cy.get('input[type="text"]').type('8.8.8.8');
    cy.get('button[type="submit"]').click();
    cy.get('pre').should('exist');
  });

  it('should allow submitting WHOIS', () => {
    cy.visit('/network');
    cy.get('#network-tool-select').select('whois');
    cy.get('input[type="text"]').type('google.com');
    cy.get('button[type="submit"]').click();
    cy.get('pre').should('exist');
  });
});
